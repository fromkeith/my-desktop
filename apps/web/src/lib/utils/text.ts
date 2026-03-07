const regexNumFind = /&#(\d+);/g;

export function unnumerializeText(text: string): string {
    return text.replace(regexNumFind, function (_, target) {
        return String.fromCharCode(target);
    });
}
