<script lang="ts">
    import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
    import { Checkbox } from "$lib/components/ui/checkbox/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import ArchiveIcon from "@lucide/svelte/icons/archive";
    import DeleteIcon from "@lucide/svelte/icons/delete";
    import MailOpenIcon from "@lucide/svelte/icons/mail-open";
    import MailIcon from "@lucide/svelte/icons/mail";

    let isHovering: boolean = false;
    let hideTimeout;

    function mouseOver() {
        clearTimeout(hideTimeout);
        isHovering = true;
    }
    function mouseOut() {
        clearTimeout(hideTimeout);
        hideTimeout = setTimeout(() => {
            isHovering = false;
        }, 100);
    }
</script>

<div
    class="flex justify-center items-center p-2 relative"
    on:mouseover={mouseOver}
    on:mouseout={mouseOut}
>
    <Checkbox />
    {#if isHovering}
        <div class="absolute left-8 top-0">
            <ButtonGroup.Root>
                <Button variant="outline">
                    <ArchiveIcon />
                </Button>
                <Button variant="outline">
                    <DeleteIcon />
                </Button>
                <Button variant="outline">
                    <MailOpenIcon />
                </Button>
                <Button variant="outline">
                    <MailIcon />
                </Button>
            </ButtonGroup.Root>
        </div>
    {/if}
</div>
