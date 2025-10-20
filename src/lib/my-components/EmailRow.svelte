<script lang="ts">
    import EmailRowActions from "$lib/my-components/EmailRowActions.svelte";
    export let isRead = false;
    let sender = "Amazon.ca";
    let preheader = 'Shipped: "Repel Umbrella Windproof..." and 3 more items';
    let receivedAt = "Oct 17";
</script>

<div class="flex w-full" class:opacity-70={isRead}>
    <EmailRowActions />
    <div class="flex grow pv-1 items-center">
        <div class="w-32 overflow-hidden" class:font-bold={!isRead}>
            {sender}
        </div>
        <div class="grow overflow-hidden">{preheader}</div>
    </div>
    <div class="w-32 overflow-hidden text-right pv-1">
        {receivedAt}
    </div>
</div>
