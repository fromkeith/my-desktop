<script lang="ts">
    import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import { Tipex } from "@friendofsvelte/tipex";
    import TipexControls from "$lib/my-components/TipexControls.svelte";
    import * as Field from "$lib/components/ui/field/index.js";
    import { Input } from "$lib/components/ui/input/index.js";
    import SendIcon from "@lucide/svelte/icons/send";
    import SaveIcon from "@lucide/svelte/icons/save";
    import TrashIcon from "@lucide/svelte/icons/trash";
    import { type IGmailEntryBody, type IComposeEmailMeta } from "$lib/models";
    import { createForm } from "felte";
    import { ZodError } from "zod";
    import EmailListInput from "$lib/my-components/EmailListInput.svelte";

    import { onMount } from "svelte";
    import {
        composeFormSchema,
        type ComposeEmailSchema,
    } from "$lib/my-components/composeEmailSchema";

    const {
        previousContents,
        srcEmail,
    }: {
        previousContents?: IGmailEntryBody | undefined;
        srcEmail: IComposeEmailMeta;
    } = $props();

    let showCc = $state(false);
    let showBcc = $state(false);

    // there are more store that could help with submitting
    let { form, data, errors, validate } = createForm({
        initialValues: {
            to: srcEmail.to ?? [],
            subject: srcEmail.subject ?? "",
            cc: srcEmail.cc ?? [],
            bcc: srcEmail.bcc ?? [],
        },
        schema: composeFormSchema,
        onSubmit: async (values: ComposeEmailSchema) => {
            console.log("onSubmit", values);
            throw new Error("no submit plz");
        },
        transform: (values: Record<string, any>) => {
            let { to, cc, bcc, ...rest } = values;
            if (typeof to !== "string") {
                to = JSON.stringify(to);
            }
            if (typeof cc !== "string") {
                cc = JSON.stringify(cc);
            }
            if (typeof bcc !== "string") {
                bcc = JSON.stringify(bcc);
            }
            return {
                ...rest,
                to: to,
                cc: cc,
                bcc: bcc,
            };
        },
        async validate(values) {
            const schema: ComposeEmailSchema = composeFormSchema;
            let { to, cc, bcc, ...rest } = values;
            if (typeof to === "string") {
                to = JSON.parse(to);
            }
            if (typeof cc === "string") {
                cc = JSON.parse(cc);
            }
            if (typeof bcc === "string") {
                bcc = JSON.parse(bcc);
            }
            const asObj: Record<string, any> = {
                ...rest,
                to,
                cc,
                bcc,
            };
            const result = await schema.safeParseAsync(asObj);
            if (result.success) {
                return {};
            }
            const err: ZodError = result.error;
            const errorResults: Record<string, string> = {};
            for (const issue of err.issues) {
                const fieldName: string = issue.path[0];
                errorResults[fieldName] = issue.message;
                if (issue.path.length > 1) {
                    errorResults[fieldName] = JSON.stringify(issue);
                }
            }
            return errorResults;
        },
    });

    const body = $derived.by(() => {
        if (previousContents) {
            return `<p></p><br/><br/><blockquote>${previousContents.html ?? previousContents.plainText}</blockquote>`;
        }
        return "";
    });

    function trySend() {
        validate().then((result) => {
            if (result.success) {
                // Send email logic here
            }
        });
    }
    function toggleCc(e: PointerEvent) {
        // for some reason... a click event is being triggered
        // when we hit enter on the EmailListInput
        if (e.pointerId === -1) {
            return;
        }
        showCc = !showCc;
    }
</script>

<div class="h-full flex flex-col">
    <form use:form>
        <Field.Group class="gap-1 mb-1">
            <Field.Field orientation="horizontal">
                <EmailListInput
                    aria-label="Add To Email Address"
                    contacts={$data.to}
                    errors={$errors.to}
                    placeholder="Send To"
                    name="to"
                />
            </Field.Field>
            {#if !showCc || !showBcc}
                <div class="text-sm text-gray-500 text-right">
                    Add
                    {#if !showCc}
                        <button
                            class="underline text-blue-600 hover:text-blue-800 inline-block"
                            onclick={toggleCc}>CC</button
                        >
                    {/if}
                    {#if !showBcc}
                        <button
                            class="underline text-blue-600 hover:text-blue-800 inline-block"
                            onclick={() => (showBcc = true)}>BCC</button
                        >
                    {/if}
                </div>
            {/if}

            {#if showCc}
                <Field.Field orientation="horizontal">
                    <EmailListInput
                        aria-label="Add CC Email Address"
                        contacts={$data.cc}
                        errors={$errors.cc}
                        name="cc"
                        placeholder="Cc"
                    />
                </Field.Field>
            {/if}
            {#if showBcc}
                <Field.Field orientation="horizontal">
                    <EmailListInput
                        contacts={$data.bcc}
                        name="bcc"
                        errors={$errors.bcc}
                        placeholder="Bcc"
                    />
                </Field.Field>
            {/if}
            <Field.Field orientation="horizontal">
                <div
                    class="border-input bg-background selection:bg-primary dark:bg-input/30 selection:text-primary-foreground ring-offset-background placeholder:text-muted-foreground shadow-xs flex min-h-9 w-full min-w-0 rounded-md border px-3 py-1 text-base outline-none transition-[color,box-shadow] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive flex-wrap cursor-text"
                >
                    <input
                        id="subject"
                        name="subject"
                        class="border-none outline-0 focus:outline-0 w-full"
                        placeholder="Subject"
                    />
                    {#if $errors.subject}
                        <div class="text-red-500">
                            {$errors.subject}
                        </div>
                    {/if}
                </div>
            </Field.Field>
        </Field.Group>
    </form>
    <Tipex {body} floating class="grow email-tipex-editor">
        {#snippet controlComponent(tipex)}
            <TipexControls {tipex}>
                <!-- Built-in utilities: copy HTML and link editing with clipboard integration -->
                <!-- Additional utility buttons can be added here -->
                <button
                    class="tipex-edit-button tipex-button-rigid"
                    aria-label="Custom action"
                >
                    hello
                </button>
            </TipexControls>
        {/snippet}
    </Tipex>
    <div class="mt-1 flex justify-end">
        <ButtonGroup.Root>
            <Button
                variant="outline"
                class="hover:text-red-500 transition-colors duration-300"
            >
                <TrashIcon />
            </Button>
            <Button
                variant="outline"
                class="hover:text-green-500 transition-colors duration-300"
            >
                <SendIcon />
                Send
            </Button>
        </ButtonGroup.Root>
    </div>
</div>
