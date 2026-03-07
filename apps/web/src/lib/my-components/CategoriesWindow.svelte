<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import { categoryListProvider } from "$lib/pods/CategoriesPod";
    import { type IWindow, WindowType } from "$lib/models";
    import { windowProvider } from "$lib/pods/WindowsPod";

    const {
        window,
    }: {
        window: IWindow;
    } = $props();

    let cats = categoryListProvider();

    function openCategory(category: string) {
        windowProvider().open({
            type: WindowType.EmailList,
            props: {
                title: category,
                filter: {
                    categories: [category],
                },
            },
        });
    }
</script>

<Window {window} title="Categories">
    {#snippet content()}
        <div>
            {#each $cats as cat (cat.category)}
                <button class="block" onclick={() => openCategory(cat.category)}
                    >({cat.messageCount}) {cat.category}</button
                >
            {/each}
        </div>
    {/snippet}
</Window>
