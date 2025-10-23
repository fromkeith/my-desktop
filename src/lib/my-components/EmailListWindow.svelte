<script lang="ts">
    import { emailListProvider } from "$lib/pods/EmailListPod";
    import Window from "$lib/my-components/Window.svelte";
    import EmailRow from "$lib/my-components/EmailRow.svelte";
    import MailsIcon from "@lucide/svelte/icons/mails";

    import type { IWindow } from "$lib/models";

    export let window: IWindow;

    $: emails = emailListProvider();
</script>

<Window {window}>
    <MailsIcon slot="window-top-left" />
    <div slot="content">
        {#each $emails as email (email.messageId)}
            <EmailRow isRead={false} {email} />
        {/each}
    </div>
</Window>
