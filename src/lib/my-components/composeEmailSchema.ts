import { z } from "zod";

const personSchema = z
    .object({
        email: z.email(),
        name: z
            .string()
            .max(512, { message: "Maximum of 512 characters allowed" })
            .optional(),
    })
    .describe("A person's email and name");

export const composeFormSchema = z.object({
    to: z
        .array(personSchema)
        .max(256, { message: "Maximum of 256 recipients allowed" })
        .min(1, { message: "At least one recipient is required" })
        .describe("The recipient's email address"),
    subject: z
        .string()
        .max(1024, { message: "Maximum of 1024 characters allowed" })
        .describe("The email subject"),
    cc: z
        .array(personSchema)
        .max(256, { message: "Maximum of 256 recipients allowed" })
        .describe("The cc's email addresses"),
    bcc: z
        .array(personSchema)
        .max(256, { message: "Maximum of 256 recipients allowed" })
        .describe("The bcc's email addresses"),
});

export type ComposeEmailSchema = typeof composeFormSchema;
