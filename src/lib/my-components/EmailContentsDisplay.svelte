<script lang="ts">
    import EmailRowActions from "$lib/my-components/EmailRowActions.svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import { WindowType, type IGmailEntryBody, type IWindow } from "$lib/models";
    import { dateFormat } from "$lib/pods/EmailListPod";
    import { emailContentsProvider } from "$lib/pods/EmailContentsPod";
    import DOMPurify from 'dompurify';

    export let messageId: string;

    $: contents = emailContentsProvider(messageId);

    $: safeContents = getSafeContents($contents);

    function getSafeContents(content: IGmailEntryBody | null): string {
        if (content === null) {
            return '';
        }
        if (content.html) {
            return DOMPurify.sanitize(content.html!);
        }
        if (content.plainText) {
            const plain = DOMPurify.sanitize(content.plainText!.trim());
            return `<p class="whitespace-pre">${plain}</p>`;
        }
        return '';
    }


</script>

<div class="border-1 overflow-auto w-full">
    {@html safeContents}
</div>
