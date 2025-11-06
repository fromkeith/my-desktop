<script lang="ts">
    import InboxIcon from "@lucide/svelte/icons/inbox";
    import StoreIcon from "@lucide/svelte/icons/store";
    import DesktopIcon from "$lib/my-components/DesktopIcon.svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import { WindowType, ComposeType } from "$lib/models";

    function openInbox() {
        windowProvider().open({
            type: WindowType.EmailList,
            props: {
                title: "Inbox",
            },
        });
    }
    function openPromotions() {
        windowProvider().open({
            type: WindowType.EmailList,
            props: {
                title: "Promotions",
                labels: ["CATEGORY_PROMOTIONS"],
            },
        });
    }
</script>

<DesktopIcon name="Inbox" onclick={openInbox}>
    {#snippet icon()}
        <InboxIcon class="size-12" />
    {/snippet}
</DesktopIcon>

<DesktopIcon name="Promotions" onclick={openPromotions}>
    {#snippet icon()}
        <StoreIcon class="size-12" />
    {/snippet}
</DesktopIcon>
