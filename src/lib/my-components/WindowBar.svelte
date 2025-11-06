<script lang="ts">
    import { Separator } from "$lib/components/ui/separator/index.js";
    import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import MinusIcon from "@lucide/svelte/icons/minus";
    import SquareIcon from "@lucide/svelte/icons/square";
    import XIcon from "@lucide/svelte/icons/x";
    import { getContext } from "svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";

    import { createEventDispatcher } from "svelte";

    export let x = 0;
    export let y = 0;
    export let title: string | undefined;

    let myWindow = getContext("window");

    const dispatch = createEventDispatcher();

    function minimize(e) {
        dispatch("minimize");
    }
    function maximize() {
        dispatch("maximize");
    }
    function close() {
        windowProvider().close(myWindow.windowId);
    }

    let root: HTMLElement;

    // drag state
    let startX = 0,
        startY = 0;
    let offsetX = 0,
        offsetY = 0;
    let dragging = false;

    function onPointerDown(e: PointerEvent) {
        if (e.target.tagName === "BUTTON") {
            return;
        }
        // where the pointer is relative to the box
        const rect = root.getBoundingClientRect();
        offsetX = e.clientX - rect.left;
        offsetY = e.clientY - rect.top;
        console.log(root);

        startX = e.clientX;
        startY = e.clientY;
        dragging = true;

        root.setPointerCapture(e.pointerId);
        e.preventDefault(); // prevent text selection
    }

    function onPointerMove(e: PointerEvent) {
        if (!dragging) return;

        // new desired top-left in viewport coords
        const desiredLeft = e.clientX - offsetX;
        const desiredTop = e.clientY - offsetY;

        // clamp to container bounds
        const c = document.body.getBoundingClientRect();
        const b = root.getBoundingClientRect(); // for width/height
        const left = Math.min(Math.max(desiredLeft, c.left), c.right);
        const top = Math.min(Math.max(desiredTop, c.top), c.bottom);

        let x = left - c.left;
        let y = top - c.top;
        dispatch("move", { x, y });
    }
    function onPointerUp(e: PointerEvent) {
        dragging = false;
        root.releasePointerCapture(e.pointerId);
    }
</script>

<div
    bind:this={root}
    on:pointermove={onPointerMove}
    on:pointerup={onPointerUp}
    on:pointercancel={onPointerUp}
    on:pointerdown={onPointerDown}
    class="overflow-hidden w-full"
>
    <div class="flex m-2 box-border">
        <slot name="window-top-left" />
        <div
            class="text-lg ml-2 overflow-hidden text-ellipsis whitespace-nowrap"
        >
            {#if title}{title}{/if}
        </div>
        <div class="grow"></div>
        <ButtonGroup.Root>
            <Button onclick={minimize} variant="outline">
                <MinusIcon />
            </Button>
            <Button onclick={maximize} variant="outline">
                <SquareIcon />
            </Button>
            <Button
                onclick={close}
                variant="outline"
                class="hover:bg-red-500 hover:text-white transition-colors duration-300"
            >
                <XIcon />
            </Button>
        </ButtonGroup.Root>
    </div>
    <Separator />
</div>
