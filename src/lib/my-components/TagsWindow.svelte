<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import { tagListProvider } from "$lib/pods/TagsPod";
    import { type IWindow, WindowType } from "$lib/models";
    import { windowProvider } from "$lib/pods/WindowsPod";

    const {
        window,
    }: {
        window: IWindow;
    } = $props();

    let cats = tagListProvider();

    function openTag(tag: string) {
        windowProvider().open({
            type: WindowType.EmailList,
            props: {
                title: tag,
                filter: {
                    tags: [tag],
                },
            },
        });
    }
</script>

<Window {window} title="Tags">
    {#snippet content()}
        <div>
            {#each $cats as cat (cat.tag)}
                <button class="block" onclick={() => openTag(cat.tag)}
                    >({cat.messageCount}) {cat.tag}</button
                >
            {/each}
        </div>
    {/snippet}
</Window>
