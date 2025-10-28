<script lang="ts">
    import { type IGmailEntryBody } from "$lib/models";
    import { emailContentsProvider } from "$lib/pods/EmailContentsPod";

    export let messageId: string;

    $: contents = emailContentsProvider(messageId);

    $: safeContents = getSafeContents($contents);

    function getSafeContents(content: IGmailEntryBody | null): string {
        if (content === null) {
            return "";
        }
        if (content.html) {
            return content.html!;
        }
        if (content.plainText) {
            return `<p class="whitespace-pre">${content.plainText!}</p>`;
        }
        return "";
    }
</script>

<div class="border-1 overflow-auto w-full">
    {@html safeContents}
</div>
