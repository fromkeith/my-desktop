<script lang="ts">
    import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import { Tipex } from "@friendofsvelte/tipex";
    import * as Field from "$lib/components/ui/field/index.js";
    import { Input } from "$lib/components/ui/input/index.js";
    import SendIcon from "@lucide/svelte/icons/send";
    import SaveIcon from "@lucide/svelte/icons/save";
    import TrashIcon from "@lucide/svelte/icons/trash";

    let body = `<p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>
        <p>The initial html content.</p>`;
</script>

<div class="h-full flex flex-col">
    <Field.Group class="gap-1 mb-1">
        <Field.Field orientation="horizontal">
            <Field.Label for="to" class="min-w-12">To</Field.Label>
            <Input id="to" required />
        </Field.Field>
        <Field.Field orientation="horizontal">
            <Field.Label for="cc" class="min-w-12">Cc</Field.Label>
            <Input id="cc" />
        </Field.Field>
        <Field.Field orientation="horizontal">
            <Field.Label for="bcc" class="min-w-12">Bcc</Field.Label>
            <Input id="bcc" />
        </Field.Field>
        <Field.Field orientation="horizontal">
            <Field.Label for="subject" class="min-w-12">Subject</Field.Label>
            <Input id="subject" required />
        </Field.Field>
    </Field.Group>
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
