<script lang="ts">
    import { ScrollArea } from "$lib/components/ui/scroll-area/index.js";
    import * as Card from "$lib/components/ui/card/index.js";
    import WindowBar from "$lib/my-components/WindowBar.svelte";

    let width = 500;
    let height = 500;
    let x = 0;
    let y = 0;

    function move(e: CustomEvent) {
        console.log("moved", e.detail);
        x = e.detail.x;
        y = e.detail.y;
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

        console.log({ rect, x: e.clientX, y: e.clientY });
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
    }
    function onPointerUp(e: PointerEvent) {
        dragging = false;
        resizebox.releasePointerCapture(e.pointerId);
    }
</script>

<div
    class="absolute"
    style:width={`${width}px`}
    style:height={`${height}px`}
    style:left={`${x}px`}
    style:top={`${y}px`}
>
    <div
        class="relative w-full h-full cursor-nwse-resize p-1"
        bind:this={resizebox}
        on:pointermove={onPointerMove}
        on:pointerup={onPointerUp}
        on:pointercancel={onPointerUp}
        on:pointerdown={onPointerDown}
    >
        <Card.Root class="overflow-hidden pt-0 h-full cursor-auto">
            <Card.Header class="px-1">
                <WindowBar on:move={move} {x} {y} />
            </Card.Header>

            <Card.Content class="overflow-hidden">
                <ScrollArea class="h-full">
                    <slot name="content" />
                </ScrollArea>
            </Card.Content>
        </Card.Root>
    </div>
</div>
