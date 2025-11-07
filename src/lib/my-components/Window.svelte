<script lang="ts">
    import { ScrollArea } from "$lib/components/ui/scroll-area/index.js";
    import * as Card from "$lib/components/ui/card/index.js";
    import WindowBar from "$lib/my-components/WindowBar.svelte";
    import type { IWindow } from "$lib/models";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import { setContext, type Snippet } from "svelte";
    import { createDebounce } from "$lib/utils/debounce";

    let {
        window,
        scrollable = true,
        title = "",
        content,
        windowTopLeft,
    }: {
        window: IWindow;
        scrollable?: boolean;
        title: string | undefined;
        content: Snippet;
        windowTopLeft?: Snippet | undefined;
    } = $props();

    setContext("window", window);

    let width = $state(window.width);
    let height = $state(window.height);
    let x = $state(window.x);
    let y = $state(window.y);

    let moveDebounce = createDebounce();

    function move(a: number, b: number) {
        x = a;
        y = b;
        moveDebounce(1000)
            .then(() => {
                windowProvider().updateWindowDim(
                    window.windowId,
                    x,
                    y,
                    width,
                    height,
                );
            })
            .catch(() => 0);
    }

    // resize state
    let lastX = 0,
        lastY = 0;
    let adjustWidth = false;
    let adjustXPos = false;
    let adjustHeight = false;
    let adjustYPos = false;
    let resizebox: HTMLElement;
    let dragging: boolean = false;
    function onPointerDown(e: PointerEvent) {
        if (e.target != resizebox) {
            return;
        }
        // where the pointer is relative to the box
        const rect = resizebox.getBoundingClientRect();

        if (e.clientX < rect.left + 16) {
            adjustWidth = true;
            adjustXPos = true;
        } else if (e.clientX > rect.right - 16) {
            adjustWidth = true;
            adjustXPos = false;
        } else {
            adjustWidth = false;
        }
        if (e.clientY < rect.top + 16) {
            adjustHeight = true;
            adjustYPos = true;
        } else if (e.clientY > rect.bottom - 16) {
            adjustHeight = true;
            adjustYPos = false;
        } else {
            adjustHeight = false;
        }

        lastX = e.clientX;
        lastY = e.clientY;
        dragging = true;

        resizebox.setPointerCapture(e.pointerId);
        e.preventDefault(); // prevent text selection
    }
    function onPointerMove(e: PointerEvent) {
        if (!dragging) return;

        let deltaX = adjustWidth ? e.clientX - lastX : 0;
        let deltaY = adjustHeight ? e.clientY - lastY : 0;

        if (adjustXPos) {
            x += deltaX;
            deltaX *= -1;
        }
        if (adjustYPos) {
            y += deltaY;
            deltaY *= -1;
        }
        width += deltaX;
        height += deltaY;
        lastX = e.clientX;
        lastY = e.clientY;

        moveDebounce(1000)
            .then(() => {
                windowProvider().updateWindowDim(
                    window.windowId,
                    x,
                    y,
                    width,
                    height,
                );
            })
            .catch(() => 0);
    }
    function onPointerUp(e: PointerEvent) {
        dragging = false;
        resizebox.releasePointerCapture(e.pointerId);
    }

    // start zindex at 100
    let zIndex = $derived(window.zIndex + 100);
    function didClick() {
        windowProvider().moveToTop(window.windowId);
    }
</script>

<div
    class="absolute"
    style:width={`${width}px`}
    style:height={`${height}px`}
    style:left={`${x}px`}
    style:top={`${y}px`}
    style:z-index={window.zIndex}
    onpointerdown={didClick}
>
    <div
        class="relative w-full h-full cursor-nwse-resize p-1"
        bind:this={resizebox}
        onpointermove={onPointerMove}
        onpointerup={onPointerUp}
        onpointercancel={onPointerUp}
        onpointerdown={onPointerDown}
    >
        <Card.Root class="overflow-hidden pt-0 h-full cursor-auto">
            <Card.Header class="px-1">
                <WindowBar onmove={move} {x} {y} {title}>
                    {#snippet windowTopLeft()}
                        {@render windowTopLeft?.()}
                    {/snippet}
                </WindowBar>
            </Card.Header>

            <Card.Content class="overflow-hidden h-full">
                {#if scrollable}
                    <ScrollArea class="h-full">
                        {@render content()}
                    </ScrollArea>
                {:else}
                    {@render content()}
                {/if}
            </Card.Content>
        </Card.Root>
    </div>
</div>
