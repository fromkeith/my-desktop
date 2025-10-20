<script lang="ts">
    import { Separator } from "$lib/components/ui/separator/index.js";
    import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import ArchiveIcon from "@lucide/svelte/icons/archive";
    import DeleteIcon from "@lucide/svelte/icons/delete";
    import MailIcon from "@lucide/svelte/icons/mail";
    import ReplyIcon from "@lucide/svelte/icons/reply";
    import ReplyAllIcon from "@lucide/svelte/icons/reply-all";
    import ForwardIcon from "@lucide/svelte/icons/forward";

    let sender = "Amazon.ca";
    let senderEmail = "no-reply@amazon.ca";
    let subject = 'Shipped: "Repel Umbrella Windproof..." and 3 more items';
    let preheader =
        "package was shipped!  Out for delivery Delivered Arriving tomorrow";
    let receivedAt = "Oct 17";
    let to = "Keith Tester";
    let toEmail = "keith@tester.com";
</script>

<h1 class="text-lg">{subject}</h1>
<div class="mb-2">
    <div class="text-md">
        <span class="font-bold">From</span>
        <span>{sender} &lt;{senderEmail}&gt;</span>
    </div>
    <div class="text-sm">
        <span class="font-bold">To</span>
        <span>{to} &lt;{toEmail}&gt;</span>
    </div>
    <div class="mt-1 flex flex-wrap">
        <ButtonGroup.Root>
            <Button variant="outline">
                <ArchiveIcon />
            </Button>
            <Button variant="outline">
                <DeleteIcon />
            </Button>
            <Button variant="outline">
                <MailIcon />
            </Button>
        </ButtonGroup.Root>
        <div class="grow"></div>
        <Button variant="outline">
            <ForwardIcon />
        </Button>
        <div class="ml-1"></div>
        <ButtonGroup.Root>
            <Button variant="outline">
                <ReplyIcon />
            </Button>
            <Button variant="outline">
                <ReplyAllIcon />
            </Button>
        </ButtonGroup.Root>
    </div>
</div>
<Separator />
<div>
    <iframe />
</div>
