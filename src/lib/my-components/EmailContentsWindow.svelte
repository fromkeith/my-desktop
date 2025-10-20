<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import EmailContents from "$lib/my-components/EmailContents.svelte";
    import type { IWindow } from "$lib/models";
    import MailOpenIcon from "@lucide/svelte/icons/mail-open";

    export let window: IWindow;
</script>

<Window {window}>
    <MailOpenIcon slot="window-top-left" />
    <EmailContents slot="content" />
</Window>
