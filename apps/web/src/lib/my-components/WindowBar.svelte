<script lang="ts">
    import { Separator } from "$lib/components/ui/separator/index.js";
    import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import MinusIcon from "@lucide/svelte/icons/minus";
    import SquareIcon from "@lucide/svelte/icons/square";
    import XIcon from "@lucide/svelte/icons/x";
    import { getContext, type Snippet } from "svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import type { IWindow } from "$lib/models";

    import { createEventDispatcher } from "svelte";

    let {
        x = 0,
        y = 0,
        title = "",
        windowTopLeft = undefined,
        onmove,
        onminimize,
        onmaximize,
    }: {
        x: number;
        y: number;
        title: string;
        windowTopLeft: Snippet | undefined;
        onmove: (x: number, y: number) => void;
        onminimize?: () => void | undefined;
        onmaximize?: () => void | undefined;
    } = $props();

    let myWindow: IWindow = getContext("window");

    function minimize() {
        onminimize?.();
    }
    function maximize() {
        onmaximize?.();
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
        if (e.target?.tagName === "BUTTON") {
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
        onmove?.(x, y);
    }
    function onPointerUp(e: PointerEvent) {
        dragging = false;
        root.releasePointerCapture(e.pointerId);
    }
</script>

<div
    bind:this={root}
    onpointermove={onPointerMove}
    onpointerup={onPointerUp}
    onpointercancel={onPointerUp}
    onpointerdown={onPointerDown}
    class="overflow-hidden w-full"
>
    <div class="flex m-2 box-border">
        {@render windowTopLeft?.()}
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
