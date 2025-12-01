<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import { categoryListProvider } from "$lib/pods/CategoriesPod";
    import type { IWindow } from "$lib/models";

    const {
        window,
    }: {
        window: IWindow;
    } = $props();

    let cats = categoryListProvider();
</script>

<Window {window} title="Categories">
    {#snippet content()}
        <div>
            {#each $cats as cat (cat.category)}
                <div>({cat.messageCount}) {cat.category}</div>
            {/each}
        </div>
    {/snippet}
</Window>
