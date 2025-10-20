<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import EmailContents from "$lib/my-components/EmailContents.svelte";
    import type { IWindow } from "$lib/models";

    export let window: IWindow;
</script>

<Window {window}>
    <EmailContents slot="content" />
</Window>
