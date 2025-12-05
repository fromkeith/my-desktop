<script lang="ts">
    import { type IGmailEntryBody } from "$lib/models";
    import { emailContentsProvider } from "$lib/pods/EmailContentsPod";

    let {
        messageId,
    }: {
        messageId: string;
    } = $props();

    let contents = emailContentsProvider(messageId);

    let safeContents = $derived(getSafeContents($contents));

    let iden = $derived(`${messageId}.${Math.random()}`);

    function unicodeToBase64(str: string) {
        const encoder = new TextEncoder();
        const data = encoder.encode(str);
        const binaryString = String.fromCharCode(...data);
        const base64 = btoa(binaryString);
        return base64;
    }

    function sizeHelper(): string {
        // weird with the scripts to svelte compiles.
        return (
            "<script" +
            `>
            let origin = '${window.location.origin}';
            let iden = '${iden}';
            function sendHeightToParent() {
            console.log('sendHeightToParent called', origin);
                    const height = document.body.scrollHeight;
                    window.parent.postMessage({ height: height, type: 'resize', iden: iden }, origin);
                }
                window.onload = sendHeightToParent;
                window.onresize = sendHeightToParent;
        </` +
            "script>"
        );
    }

    function getSafeContents(content: IGmailEntryBody | null): string {
        if (content === null) {
            return `data:text/html;charset=utf-8;base64,${unicodeToBase64(sizeHelper())}`;
        }
        if (content.html) {
            return `data:text/html;charset=utf-8;base64,${unicodeToBase64(sizeHelper() + content.html!)}`;
        }
        if (content.plainText) {
            const text = `<p class="whitespace-pre">${content.plainText!}</p>`;
            return `data:text/html;charset=utf-8;base64,${unicodeToBase64(sizeHelper() + text)}`;
        }
        return "";
    }

    let iframeHeight = $state(200);

    function windowMessage(event: MessageEvent) {
        console.log("Received message:", event.data);
        if (event.data.type === "resize" && event.data.iden === iden) {
            const height = event.data.height;
            iframeHeight = Math.min(Math.max(height, 200), window.innerHeight);
            console.log("iframeHeight updated", iframeHeight);
        }
    }
</script>

<svelte:window onmessage={windowMessage} />

<div class="border-1 overflow-auto w-full p-4">
    {#if safeContents}
        <iframe
            class="w-full h-full overflow-auto"
            style:height={`${iframeHeight}px`}
            src={safeContents}
        ></iframe>
    {/if}
</div>
