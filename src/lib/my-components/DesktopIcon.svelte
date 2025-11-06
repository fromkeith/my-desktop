<script lang="ts">
    import type { Snippet } from "svelte";
    let {
        name,
        icon,
        onclick,
    }: {
        name: string;
        icon: Snippet;
        onclick: (e: MouseEvent) => void;
    } = $props();
</script>

<div class="mt-15 ml-4">
    <button
        {onclick}
        ondoubleclick={() => void 0}
        aria-label={name}
        class="w-18 h-18 flex flex-col justify-center items-center hover:border-gray-400 hover:shadow hover:bg-white border-transparent p-2 rounded border-2 box-border cursor-pointer"
    >
        {@render icon()}
        <span class="text-sm">{name}</span>
    </button>
</div>
