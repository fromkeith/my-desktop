<script lang="ts">
    import StoreIcon from "@lucide/svelte/icons/store";
    import NewsPaperIcon from "@lucide/svelte/icons/newspaper";

    let {
        labels,
    }: {
        labels?: string[];
    } = $props();

    let labelSet = $derived(new Set(labels ?? []));
    let isInStore = $derived(labelSet.has("CATEGORY_PROMOTIONS"));
    let isUpdate = $derived(labelSet.has("CATEGORY_UPDATES"));
</script>

{#if isInStore}
    <div title="Promotions" class="opacity-80 mr-1 mt-1">
        <StoreIcon class="size-3" />
    </div>
{/if}
{#if isUpdate}
    <div title="Updates" class="opacity-80 mr-1 mt-1">
        <NewsPaperIcon class="size-3" />
    </div>
{/if}
