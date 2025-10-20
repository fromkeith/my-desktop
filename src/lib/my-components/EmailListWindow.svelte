<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import EmailRow from "$lib/my-components/EmailRow.svelte";

    import type { IWindow } from "$lib/models";

    export let window: IWindow;
</script>

<Window {window}>
    <div slot="content">
        <EmailRow isRead={false} />
        <EmailRow isRead={true} />
        <EmailRow isRead={true} />
        <EmailRow />
        <EmailRow isRead={true} />
    </div>
</Window>
