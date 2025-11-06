<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import EmailThread from "$lib/my-components/EmailThread.svelte";
    import type { IWindow, IGmailEntry } from "$lib/models";
    import MailOpenIcon from "@lucide/svelte/icons/mail-open";

    export let window: IWindow;
    export let email: IGmailEntry;
</script>

<Window {window} title={email.subject}>
    <MailOpenIcon slot="window-top-left" />
    <EmailThread slot="content" {email} />
</Window>
