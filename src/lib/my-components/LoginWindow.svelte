<script lang="ts">
    import type { IWindow } from "$lib/models";
    import Window from "$lib/my-components/Window.svelte";

    import OAuth from "./OAuth.svelte";

    let window: IWindow = {
        zIndex: 0,
        width: 400,
        height: 300,
        x: Math.max(document.body.clientWidth / 2 - 400 / 2, 0),
        y: Math.max(document.body.clientHeight / 2 - 400 / 2, 0),
    };
</script>

<Window {window}>
    <div slot="content">
        <p>Login to My Desktop</p>
        <OAuth />
    </div>
</Window>
