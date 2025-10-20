<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import ComposeEmailContents from "$lib/my-components/ComposeEmailContents.svelte";
    import MailPlusIcon from "@lucide/svelte/icons/mail-plus";

    import type { IWindow } from "$lib/models";

    export let window: IWindow;
</script>

<Window {window} scrollable={false}>
    <MailPlusIcon slot="window-top-left" />
    <ComposeEmailContents slot="content" />
</Window>
