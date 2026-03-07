<script lang="ts">
    import * as Item from "$lib/components/ui/item/index.js";
    import type { IPersonInfo } from "$lib/models";
    import CopyIcon from "@lucide/svelte/icons/Copy";
    import CheckIcon from "@lucide/svelte/icons/check";
    import XIcon from "@lucide/svelte/icons/x";

    const {
        contact,
        doCopy = true,
        doClose = false,
        onclick,
        onremove,
        highlight,
    }: {
        contact: IPersonInfo;
        doCopy?: boolean;
        doClose?: boolean;
        showEmail?: boolean;
        onremove?: (c: IPersonInfo) => void;
        highlight?: { tooltip: string; class: string } | undefined;
        onclick?: () => void;
    } = $props();

    let highlightClass = $derived(highlight?.class || "");
    let contents = $derived(`${contact.name} <${contact.email}>`);
    let didCopy = $state(false);

    async function copy() {
        try {
            await navigator.clipboard.writeText(contents);
            didCopy = true;
            setTimeout(() => {
                didCopy = false;
            }, 3000);
        } catch (ex) {}
    }
</script>

<Item.Root
    class="m-1 p-1 word-break {highlightClass}"
    variant="outline"
    size="sm"
>
    <Item.Content {onclick}>
        {contents}
    </Item.Content>
    <Item.Actions>
        {#if doCopy}
            <button onclick={copy}>
                {#if didCopy}
                    <CheckIcon class="size-4 bg-green-500 text-white rounded" />
                {:else}
                    <CopyIcon class="size-4" />
                {/if}
            </button>
        {/if}
        {#if doClose}
            <button onclick={() => onremove?.(contact)}
                ><XIcon class="size-4" /></button
            >
        {/if}
    </Item.Actions>
</Item.Root>
