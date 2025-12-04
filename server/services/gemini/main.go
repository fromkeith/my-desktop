package main

import (
	"context"
	"fmt"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"fromkeith/my-desktop-server/services/kafkaservice"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/genai"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type messageBody struct {
	entry         data.GmailEntry
	body          string
	result        *expectedAnalyzeResult
	src           kafka.Message
	embedding     *genai.ContentEmbedding
	embeddingText string
}
type expectedAnalyzeResult struct {
	Theme      string
	Summary    string
	Categories []string
	Tags       []string
	Todos      []string
}

var outputDimens int32 = 3072 // its the default, but lets be explict

func main() {
	ctx := context.WithValue(context.Background(), "service", "gemini")

	available := globals.KafkaWriter("email_embedding_available")
	defer available.Close()

	kafkaservice.Run(ctx, kafkaservice.KafkaService{
		Name:        "gemini",
		Topic:       "email_injest_available",
		Group:       "gemini",
		NumMessages: 10,
		MaxWait:     time.Second,
		NumWorkers:  2,
		Worker: func(ctx context.Context, msgs []kafka.Message) (dlq []kafka.Message, err error) {
			return work(ctx, msgs, available)
		},
		Dlq: "gemini_dlq",
	})

}

func work(ctx context.Context, msgs []kafka.Message, available *kafka.Writer) (dlq []kafka.Message, err error) {

	failed := make([]kafka.Message, 0)
	bodies := make([]messageBody, 0, len(msgs))
	for _, msg := range msgs {
		log.Info().
			Ctx(ctx).
			Str("taskId", string(msg.Key)).
			Msg("processing message")
		var payload data.EmailInjestedPayload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Str("taskId", string(msg.Key)).
				Msg("failed to unmarshal email")
			failed = append(failed, msg)
			continue
		}
		entry := payload.Entry
		entry.AccountId = payload.AccountId // needed since accountId doesn't marshal to json
		body, err := fetchBody(ctx, entry)
		if err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Str("taskId", string(msg.Key)).
				Msg("failed to fetch message body")
			failed = append(failed, msg)
			continue
		}
		// TODO: strip out replies? but how do we know?
		// TODO: what about attachments?
		asText, err := stripHtml(ctx, *body)
		if err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Str("taskId", string(msg.Key)).
				Msg("failed to strip html")
			failed = append(failed, msg)
			continue
		}
		bodies = append(bodies, messageBody{
			entry: entry,
			body:  asText,
			src:   msg,
		})
	}

	// analyze each body via gemini
	for i, msg := range bodies {
		log.Info().
			Ctx(ctx).
			Str("taskId", msg.entry.ToDocumentId()).
			Int("payloadSize", len(msg.body)).
			Msg("ai-ing document")

		analyzeResult, err := anaylze(ctx, msg)
		if err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Str("taskId", msg.entry.ToDocumentId()).
				Msg("failed to fetch message")
			failed = append(failed, msg.src)
			continue
		}
		bodies[i].result = analyzeResult
	}

	// write the found tags + categories
	if err := writeAnalyzeResult(ctx, bodies); err != nil {
		log.Error().
			Ctx(ctx).
			Stack().
			Err(err).
			Msg("failed to write tags and categories")
	}

	// create the embeddings and assign them to bodies
	toSave := make([]data.EmailSummaryEmbedding, 0, len(bodies))
	if len(bodies) > 0 {
		results, err := createEmbeddings(ctx, bodies)
		if err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Msg("failed to create embeddings")
			for _, msg := range bodies {
				if msg.result != nil {
					failed = append(failed, msg.src)
				}
			}
		} else {
			// assign embeddings to bodies
			var bodyI int = 0
			for _, embd := range results {
				// we expect things to be in the right order
				// but if one failed to analyze, we need to skip it
				// as it won't have an embedding
				for bodies[bodyI].result == nil {
					bodyI++
				}
				msg := bodies[bodyI]
				msg.embedding = embd
				bodies[bodyI] = msg
				// go to next body
				bodyI++

				toSave = append(toSave, data.EmailSummaryEmbedding{
					MessageId: msg.entry.MessageId,
					AccountId: msg.entry.AccountId,
					Embedding: embd.Values,
					Sender:    msg.entry.Sender,
					Receiver:  msg.entry.Receiver,
					Summary:   msg.embeddingText,
					Version:   0,
				})
			}
		}
	}
	// save the embeddings with metadata
	if len(toSave) > 0 {
		data.BulkWriteEmailSummaries(ctx, toSave)
		nextStep := make([]kafka.Message, 0, len(toSave))
		for _, entry := range toSave {
			entryBytes, _ := json.Marshal(entry)
			nextStep = append(nextStep, kafka.Message{
				Key:   []byte(entry.ToDocumentId()),
				Value: entryBytes,
			})
		}
		// make it available to downstream services
		if err := available.WriteMessages(ctx, nextStep...); err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Msg("failed to write messages to available topic. Messages lost!")
		}
	}
	// return failed
	return failed, nil

}

var responseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"Theme": map[string]any{
			"type":        "string",
			"description": "Theme of the email",
		},
		"Summary": map[string]any{
			"type":        "string",
			"description": "1-3 line summary of this email",
		},
		"Categories": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":        "string",
				"description": "The L1 and L2 categories for this email.",
			},
		},
		"Tags": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":        "string",
				"description": "A tag for this email. Must be 1 word.",
			},
		},
		"Todos": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":        "string",
				"description": "An action item to complete when replying to this email",
			},
			"description": "A list of action items to complete when replying to this email. Can be an empty array if no action is needed.",
		},
	},
	"required": []string{"Theme", "Summary", "Categories", "Tags", "Todos"},
}

func anaylze(ctx context.Context, email messageBody) (*expectedAnalyzeResult, error) {

	result, err := globals.Gemini().Models.GenerateContent(
		ctx,
		"gemini-flash-latest",
		genai.Text(`Subject: `+email.entry.Subject+`\n\n`+email.body),
		&genai.GenerateContentConfig{
			ResponseMIMEType:   "application/json",
			ResponseJsonSchema: responseSchema,
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{
					Text: gemini_instructions,
				}},
			},
		},
	)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Str("docId", email.entry.ToDocumentId()).
			Msg("failed to generate content")
		return nil, err
	}
	txt := result.Text()
	var res expectedAnalyzeResult
	if err := json.Unmarshal([]byte(txt), &res); err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Str("docId", email.entry.ToDocumentId()).
			Msg("failed to unmarshal analyze result")
		return nil, err
	}
	log.Info().Ctx(ctx).Any("analyzeResult", res).Msg("analyzeResult")
	return &res, nil
}

func createEmbeddings(ctx context.Context, items []messageBody) ([]*genai.ContentEmbedding, error) {
	contents := make([]*genai.Content, 0, len(items))
	for i, item := range items {
		if item.result == nil {
			continue
		}
		res := item.result
		toEmbedd := fmt.Sprintf("%s\n%s\n%s\n%s", res.Theme, res.Summary, strings.Join(res.Categories, ", "), strings.Join(res.Tags, ", "))
		items[i].embeddingText = toEmbedd
		contents = append(contents, genai.NewContentFromText(toEmbedd, genai.RoleUser))
	}

	// https://ai.google.dev/gemini-api/docs/embeddings
	result, err := globals.Gemini().Models.EmbedContent(
		ctx, "gemini-embedding-001",
		contents,
		&genai.EmbedContentConfig{
			TaskType:             "CLUSTERING",
			OutputDimensionality: &outputDimens,
		},
	)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Msg("failed to create embedding")
		return nil, err
	}
	return result.Embeddings, nil
}

func fetchBody(ctx context.Context, entry data.GmailEntry) (*data.GmailEntryBody, error) {
	result := globals.DocDb().Collection("MessageBodies").FindOne(
		ctx,
		bson.M{"_id": entry.ToDocumentId()},
	)
	var body data.GmailEntryBody
	if err := result.Decode(&body); err != nil {
		return nil, err
	}
	return &body, nil
}

func stripHtml(ctx context.Context, body data.GmailEntryBody) (string, error) {
	if body.PlainText != "" {
		return strings.TrimSpace(body.PlainText), nil
	}
	if body.Html != "" {
		// niave limit of body size
		if len(body.Html) > 1024*1024*2 {
			body.Html = body.Html[:1024*1024*2]
		}
		res, err := htmltomarkdown.ConvertString(body.Html)
		return strings.TrimSpace(res), err
	}
	// empty body.. so be e
	return "empty body", nil
}

func writeAnalyzeResult(ctx context.Context, bodies []messageBody) error {
	tagsAndCategories := make([]mongo.WriteModel, 0, len(bodies))
	for _, msg := range bodies {
		if msg.result == nil {
			continue
		}
		// enforce normalization
		for i, tag := range msg.result.Tags {
			msg.result.Tags[i] = strings.TrimSpace(strings.ToLower(tag))
		}
		for i, cat := range msg.result.Categories {
			msg.result.Categories[i] = strings.TrimSpace(strings.ToLower(cat))
		}
		// add to the message itself
		addToSet := bson.M{}
		if len(msg.result.Tags) > 0 {
			addToSet["tags"] = bson.D{{"$each", msg.result.Tags}}
		}
		if len(msg.result.Categories) > 0 {
			addToSet["categories"] = bson.D{{"$each", msg.result.Categories}}
		}
		todos := bson.M{}
		if len(msg.result.Todos) > 0 {
			todos["todos"] = msg.result.Todos
		}
		if len(addToSet) == 0 && len(todos) == 0 {
			continue // nothing to add
		}
		update := bson.M{
			"$currentDate": bson.M{"updatedAt": true},
			"$inc":         bson.M{"revisionCount": 1},
		}
		if len(addToSet) > 0 {
			update["$addToSet"] = addToSet
		}
		if len(todos) > 0 {
			update["$set"] = todos
		}
		tagsAndCategories = append(tagsAndCategories, mongo.NewUpdateOneModel().
			SetFilter(bson.D{{"_id", msg.entry.ToDocumentId()}}).
			SetUpdate(update).
			SetUpsert(false),
		)
	}
	col := globals.DocDb().Collection("Messages")
	if _, err := col.BulkWrite(ctx, tagsAndCategories); err != nil {
		return err
	}
	return nil
}

// TODO: pull this from a database
// allow the user to define their own categories
const (
	predefined_categories = `
	# Accounts & Identity
- Account setup
- Profile updates
- Login activity
- Password reset
- Identity verification
- Account recovery
- Access permission changes
- Account closure

# Security & Fraud
- Security alerts
- Suspicious activity
- Fraud prevention
- Device verification
- Data breach notifications
- Authentication issues
- Privacy notices
- Security policy updates

# Billing & Payments
- Payment confirmations
- Receipts
- Invoices
- Billing errors
- Payment failures
- Refunds
- Billing plan changes
- Tax-related billing

# Finance & Banking
- Account balance updates
- Transactions
- Money transfers
- Loan and credit updates
- Investment statements
- Market advisories
- Financial reports
- Tax documents

# Shopping & Ecommerce
- Order confirmations
- Shipping updates
- Delivery notices
- Returns & refunds
- Promotions
- Loyalty rewards
- Product availability
- Online receipts

# Travel & Transportation
- Flight bookings
- Hotel reservations
- Itineraries
- Schedule changes
- Travel alerts
- Transportation services
- Travel insurance
- Visa & entry documentation

# Health & Medical
- Appointment confirmations
- Lab results
- Medical billing
- Health insurance claims
- Pharmacy & prescriptions
- Health reminders
- Telehealth communication
- Provider updates

# Employment & Hiring
- Job applications
- Recruiting outreach
- Interview scheduling
- Offer letters
- Hiring decisions
- Career development
- Background checks
- Job search resources

# Human Resources
- Onboarding materials
- Payroll notices
- Benefits updates
- HR policy updates
- Performance evaluations
- Workplace compliance
- Time-off approvals
- Employee communication

# Internal Operations
- Meeting invitations
- Project updates
- Scheduling
- Operational alerts
- Resource allocation
- Organizational announcements
- Workflow coordination
- Team communication

# Engineering & Technical
- Deployments
- Build notifications
- Incident alerts
- API changes
- Bug reports
- Logging & monitoring
- Developer tools
- System upgrades

# Product & UX
- Product updates
- Feature announcements
- Release notes
- User research
- UX testing invitations
- Product onboarding
- Churn outreach
- Usage insights

# Sales & Business Development
- Lead communication
- Sales proposals
- Pricing quotes
- Contract negotiations
- CRM updates
- Renewal discussions
- Partner outreach
- Deal progress updates

# Marketing & Growth
- Marketing campaigns
- Newsletters
- Market research
- Audience insights
- Promotional materials
- Brand partnerships
- SEO & content updates
- Performance reports

# Customer Support
- Support tickets
- Case updates
- Troubleshooting instructions
- Service notifications
- Customer follow-ups
- Satisfaction surveys
- Replacement approvals
- Escalation updates

# Leadership & Executive
- Board communication
- Investor updates
- Strategic planning
- Executive alignment
- Organizational changes
- Financial reporting
- High-level partnerships
- Crisis communication

# Legal & Compliance
- Legal notices
- Contract documents
- Compliance reminders
- Licensing updates
- Privacy rights requests
- Regulatory filings
- Policy acknowledgments
- Audit correspondence

# Government & Civic
- Government forms
- Official notices
- Civic programs
- Elections & voting
- Public safety alerts
- Permits & licensing
- Local government updates
- Public policy changes

# Education & School
- School announcements
- Teacher communication
- Academic schedules
- Assignments
- Grades & transcripts
- PTA/parent updates
- Student services
- Educational programs

# Parenting & Childcare
- Childcare scheduling
- School activities
- Family event planning
- Youth programs
- Parent groups
- Activity reminders
- Child health updates
- Permission forms

# Home & Household
- Maintenance appointments
- Home services
- Utility billing
- Property management
- Renovation updates
- Landscaping & gardening
- Home monitoring
- Household purchases

# Real Estate & Property
- Rental agreements
- Mortgage communication
- Property listings
- Home inspections
- Realtor communication
- Tenant updates
- Move-in/move-out notices
- Insurance evaluations

# Legal & Financial Services
- Accounting communication
- Insurance quotes
- Financial advising
- Estate planning
- Legal consultations
- Will & trust services
- Service agreements
- Professional recommendations

# Social & Community
- Event invitations
- RSVPs
- Group messages
- Social media notifications
- Community updates
- Friend/connection requests
- Local event notices
- Social planning

# Sports & Recreation
- Sports leagues
- Team updates
- Fitness classes
- Event schedules
- Results & scores
- Recreational programs
- Outdoor activities
- Hobby group communication

# Entertainment & Media
- Streaming updates
- Content releases
- Game updates
- Podcasts
- Music events
- Book/news alerts
- Subscriptions
- Media promotions

# Nonprofit & Charity
- Donation confirmations
- Fundraising campaigns
- Volunteering
- Charity events
- Nonprofit updates
- Impact reports
- Member communication
- Advocacy alerts

# Product Notifications & Systems
- App notifications
- System alerts
- Feature rollouts
- Service interruptions
- Device updates
- Usage warnings
- Platform maintenance
- Data export availability

# Personal Organization
- Reminders
- Notes & documents
- Personal tasks
- Calendar scheduling
- Lists & planning
- Personal updates
- Life admin
- Digital tools

# Miscellaneous
- Uncategorized
- Unknown intent
- Mixed-content messages
- Broad newsletters
- Irrelevant content
- Auto-generated noise
- Spam indicators
- System noise
	`
)

var gemini_instructions = fmt.Sprintf(`You help categorize & summarize emails. Get out the theme (1 line), a summary (1-3 lines), suggested categories (L1, L2), tags (3-8 short tokens, as a string list), and a TODO list for replying to the email.
Return as JSON (Theme, Summary, Categories, Tags, Todos)

The categories are below, L1 is the heading, L2 is bullet under the heading:

%s
`, predefined_categories)
