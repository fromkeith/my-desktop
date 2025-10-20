<script lang="ts">
    import EmailRowActions from "$lib/my-components/EmailRowActions.svelte";
    export let isRead = false;
    let sender = "Amazon.ca";
    let subject = 'Shipped: "Repel Umbrella Windproof..." and 3 more items';
    let preheader =
        "package was shipped!  Out for delivery Delivered Arriving tomorrow";
    let receivedAt = "Oct 17";
</script>

<div class="flex w-full" class:opacity-70={isRead}>
    <EmailRowActions />
    <div class="flex flex-1 pv-1 min-w-0 items-center">
        <div class="w-32 overflow-hidden" class:font-bold={!isRead}>
            {sender}
        </div>
        <div class="grow overflow-hidden min-w-0">
            <div class="truncate">{subject}</div>
            <div class="text-sm truncate">{preheader}</div>
        </div>
    </div>
    <div class="w-16 overflow-hidden text-right pv-1">
        {receivedAt}
    </div>
</div>
