<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import EmailThread from "$lib/my-components/EmailThread.svelte";
    import type { IWindow, IGmailEntry } from "$lib/models";
    import MailOpenIcon from "@lucide/svelte/icons/mail-open";
    import { emailThreadProvider } from "$lib/pods/EmailThreadPod";

    let {
        window,
        threadId,
        openTo,
    }: {
        window: IWindow;
        threadId: string;
        openTo?: string;
    } = $props();

    let thread = emailThreadProvider(threadId);

    let last = $derived(
        $thread.length > 0 ? $thread[$thread.length - 1] : null,
    );
</script>

<Window {window} title={last?.subject}>
    {#snippet windowTopLeft()}
        <MailOpenIcon />
    {/snippet}
    {#snippet content()}
        <EmailThread thread={$thread} {threadId} openMessageId={openTo} />
    {/snippet}
</Window>
