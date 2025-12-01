<script lang="ts">
    import { emailListProvider } from "$lib/pods/EmailListPod";
    import Window from "$lib/my-components/Window.svelte";
    import EmailRow from "$lib/my-components/EmailRow.svelte";
    import MailsIcon from "@lucide/svelte/icons/mails";

    import type { IWindow, IEmailListOptions } from "$lib/models";

    let {
        window,
        filter,
        title,
    }: {
        window: IWindow;
        filter?: IEmailListOptions;
        title?: string;
    } = $props();

    let emails = $derived(emailListProvider(filter));
</script>

<Window {window} {title}>
    {#snippet windowTopLeft()}
        <MailsIcon />
    {/snippet}
    {#snippet content()}
        {#each $emails as email (email.messageId)}
            <EmailRow {email} />
        {/each}
    {/snippet}
</Window>
