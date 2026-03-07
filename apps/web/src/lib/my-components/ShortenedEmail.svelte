<script lang="ts">
    import * as Tooltip from "$lib/components/ui/tooltip/index.js";
    import type { IPersonInfo } from "$lib/models";
    import EmailContact from "./EmailContact.svelte";

    const {
        contact,
        highlight,
        onclick,
    }: {
        contact: IPersonInfo;
        highlight: { tooltip: string; class: string } | undefined;
        onclick?: () => void;
    } = $props();

    let highlightClass = $derived(highlight?.class || "");
</script>

<Tooltip.Provider>
    <Tooltip.Root>
        <Tooltip.Trigger class={highlightClass} {onclick}
            >{contact.name || contact.email}</Tooltip.Trigger
        >
        <Tooltip.Content>
            {#if highlight}
                <div class={highlight.class}>
                    {highlight.tooltip}
                </div>
            {/if}
            <EmailContact {contact} doClose={false} />
        </Tooltip.Content>
    </Tooltip.Root>
</Tooltip.Provider>
