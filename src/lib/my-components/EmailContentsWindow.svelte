<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import EmailThread from "$lib/my-components/EmailThread.svelte";
    import type { IWindow, IGmailEntry } from "$lib/models";
    import MailOpenIcon from "@lucide/svelte/icons/mail-open";

    let {
        window,
        email,
    }: {
        window: IWindow;
        email: IGmailEntry;
    } = $props();
</script>

<Window {window} title={email.subject}>
    {#snippet windowTopLeft()}
        <MailOpenIcon />
    {/snippet}
    {#snippet content()}
        <EmailThread {email} />
    {/snippet}
</Window>
