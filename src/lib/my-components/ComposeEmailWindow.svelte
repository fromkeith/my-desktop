<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import ComposeEmailContents from "$lib/my-components/ComposeEmailContents.svelte";

    import type { IWindow } from "$lib/models";

    export let window: IWindow;
</script>

<Window {window} scrollable={false}>
    <ComposeEmailContents slot="content" />
</Window>
