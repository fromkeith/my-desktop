<script lang="ts">
    import { emailListProvider } from "$lib/pods/EmailListPod";
    import Window from "$lib/my-components/Window.svelte";
    import EmailRow from "$lib/my-components/EmailRow.svelte";
    import MailsIcon from "@lucide/svelte/icons/mails";

    import type { IWindow } from "$lib/models";

    let {
        window,
        labels,
        title,
    }: {
        window: IWindow;
        labels?: string[];
        title?: string;
    } = $props();

    let emails = $derived(emailListProvider(labels));
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
