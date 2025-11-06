<script lang="ts">
    import StoreIcon from "@lucide/svelte/icons/store";
    import NewsPaperIcon from "@lucide/svelte/icons/newspaper";
    import StarIcon from "@lucide/svelte/icons/star";
    import CrownIcon from "@lucide/svelte/icons/crown";
    import UserIcon from "@lucide/svelte/icons/user";
    import SpeechIcon from "@lucide/svelte/icons/speech";
    import MessageCircleIcon from "@lucide/svelte/icons/message-circle";
    import ShoppingBagIcon from "@lucide/svelte/icons/shopping-bag";

    let {
        labels,
    }: {
        labels?: string[];
    } = $props();

    let labelSet = $derived(new Set(labels ?? []));

    let icons = $derived.by(() => {
        let icons = [];
        if (labelSet.has("CATEGORY_PROMOTIONS")) {
            icons.push({
                icon: StoreIcon,
                title: "Promotions",
            });
        }
        if (labelSet.has("CATEGORY_UPDATES")) {
            icons.push({
                icon: NewsPaperIcon,
                title: "Updates",
            });
        }
        if (labelSet.has("CATEGORY_SOCIAL")) {
            icons.push({
                icon: MessageCircleIcon,
                title: "Social",
            });
        }
        if (labelSet.has("CATEGORY_FORUMS")) {
            icons.push({
                icon: SpeechIcon,
                title: "Forums",
            });
        }
        if (labelSet.has("CATEGORY_PERSONAL")) {
            icons.push({
                icon: UserIcon,
                title: "Personal",
            });
        }
        if (labelSet.has("CATEGORY_PURCHASES")) {
            icons.push({
                icon: ShoppingBagIcon,
                title: "Purchases",
            });
        }
        if (labelSet.has("STARRED")) {
            icons.push({
                icon: StarIcon,
                title: "Starred",
            });
        }
        if (labelSet.has("IMPORTANT")) {
            icons.push({
                icon: CrownIcon,
                title: "Important",
            });
        }
        return icons;
    });
</script>

{#each icons as icon}
    {@const IconComponent = icon.icon}
    <div title={icon.title} class="opacity-80 mr-1 mt-1">
        <IconComponent class="size-3" />
    </div>
{/each}
