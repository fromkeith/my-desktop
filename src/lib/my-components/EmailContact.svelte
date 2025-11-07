<script lang="ts">
    import * as Item from "$lib/components/ui/item/index.js";
    import type { IPersonInfo } from "$lib/models";
    import CopyIcon from "@lucide/svelte/icons/Copy";
    import XIcon from "@lucide/svelte/icons/x";

    const {
        contact,
        doCopy = true,
        doClose = false,
        onremove,
        highlight,
    }: {
        contact: IPersonInfo;
        doCopy?: boolean;
        doClose?: boolean;
        showEmail?: boolean;
        onremove?: (c: IPersonInfo) => void;
        highlight?: { tooltip: string; class: string } | undefined;
    } = $props();

    let highlightClass = $derived(highlight?.class || "");
</script>

<Item.Root
    class="m-1 p-1 word-break {highlightClass}"
    variant="outline"
    size="sm"
>
    <Item.Content>
        {contact.name}
        &lt;{contact.email}&gt;
    </Item.Content>
    <Item.Actions>
        {#if doCopy}
            <button><CopyIcon class="size-4" /></button>
        {/if}
        {#if doClose}
            <button onclick={() => onremove?.(contact)}
                ><XIcon class="size-4" /></button
            >
        {/if}
    </Item.Actions>
</Item.Root>
