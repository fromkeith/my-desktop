<script lang="ts">
    import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import { Tipex } from "@friendofsvelte/tipex";
    import * as Field from "$lib/components/ui/field/index.js";
    import { Input } from "$lib/components/ui/input/index.js";
    import SendIcon from "@lucide/svelte/icons/send";
    import SaveIcon from "@lucide/svelte/icons/save";
    import TrashIcon from "@lucide/svelte/icons/trash";
    import { type IGmailEntryBody, type IComposeEmailMeta } from "$lib/models";
    import { createForm } from "felte";
    import { ZodError } from "zod";

    import { onMount } from "svelte";
    import {
        composeFormSchema,
        type ComposeEmailSchema,
    } from "$lib/my-components/composeEmailSchema";

    export let previousContents: IGmailEntryBody;
    export let data: IComposeEmailMeta;

    onMount(async () => {});

    let { form } = createForm({
        initialValues: {
            to: data.to.map((a) => a.email).join(", "),
            subject: data.subject ?? "",
            cc: data.cc?.map((a) => a.email).join(", ") ?? "",
            bcc: data.bcc?.map((a) => a.email).join(", ") ?? "",
        },
        schema: composeFormSchema,
        onSubmit: async (values: ComposeEmailSchema) => {
            console.log(values);
        },
        async validate(values) {
            const schema: ComposeEmailSchema = composeFormSchema;
            const result = await schema.safeParseAsync(values);
            if (result.success) {
                return;
            }
            const err: ZodError = result.error;
            console.log(err);
        },
    });

    $: body = `<p></p><br/><br/><blockquote>${previousContents.html ?? previousContents.plainText}</blockquote>`;
</script>

<div class="h-full flex flex-col">
    <form use:form>
        <Field.Group class="gap-1 mb-1">
            <Field.Field orientation="horizontal">
                <Field.Label for="to" class="min-w-12">To</Field.Label>
                <Input id="to" name="to" required />
            </Field.Field>
            <Field.Field orientation="horizontal">
                <Field.Label for="cc" class="min-w-12">Cc</Field.Label>
                <Input id="cc" name="cc" />
            </Field.Field>
            <Field.Field orientation="horizontal">
                <Field.Label for="bcc" class="min-w-12">Bcc</Field.Label>
                <Input id="bcc" name="ccc" />
            </Field.Field>
            <Field.Field orientation="horizontal">
                <Field.Label for="subject" class="min-w-12">Subject</Field.Label
                >
                <Input id="subject" name="subject" required />
            </Field.Field>
        </Field.Group>
    </form>
    <Tipex {body} floating />
    <div class="mt-1 flex justify-end">
        <ButtonGroup.Root>
            <Button variant="outline">
                <TrashIcon />
            </Button>
            <Button variant="outline">
                <SaveIcon />
            </Button>
            <Button variant="outline">
                <SendIcon />
                Send
            </Button>
        </ButtonGroup.Root>
    </div>
</div>
