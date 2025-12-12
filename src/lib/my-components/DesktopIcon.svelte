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

<div class="m-4">
    <button
        {onclick}
        ondoubleclick={() => void 0}
        aria-label={name}
        class="w-18 h-18 flex flex-col justify-center items-center p-2 rounded cursor-pointer group"
    >
        <div
            class="bg-gray-50 rounded-full p-2 min-w-12 min-h-12 overflow-hidden group-hover:bg-blue-100 group-hover:shadow-md"
        >
            {@render icon()}
        </div>
        <span
            class="text-sm text-shadow-white text-shadow-sm group-hover:text-shadow-md"
            >{name}</span
        >
    </button>
</div>
