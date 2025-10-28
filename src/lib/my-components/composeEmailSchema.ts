import { z } from "zod";

export const composeFormSchema = z.object({
    to: z
        .string()
        .min(2)
        .max(50)
        .describe("The recipient's email address")
        .regex(/^[^\s@]+@[^\s@]+\.[^\s@]+$/),
    subject: z.string().max(1024).describe("The email subject"),
    cc: z
        .string()
        .max(50)
        .describe("The cc's email addresses")
        .regex(/^[^\s@]+@[^\s@]+\.[^\s@]+$/),
    bcc: z
        .string()
        .max(50)
        .describe("The bcc's email addresses")
        .regex(/^[^\s@]+@[^\s@]+\.[^\s@]+$/),
});

export type ComposeEmailSchema = typeof composeFormSchema;
