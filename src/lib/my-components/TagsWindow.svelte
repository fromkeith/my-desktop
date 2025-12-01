<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import { tagListProvider } from "$lib/pods/TagsPod";
    import type { IWindow } from "$lib/models";

    const {
        window,
    }: {
        window: IWindow;
    } = $props();

    let cats = tagListProvider();
</script>

<Window {window} title="Tags">
    {#snippet content()}
        <div>
            {#each $cats as cat (cat.tag)}
                <div>({cat.messageCount}) {cat.tag}</div>
            {/each}
        </div>
    {/snippet}
</Window>
