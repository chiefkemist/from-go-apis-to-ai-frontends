
import crypto from "node:crypto";

export async function POST(request: Request) {
    try {
        const formData = await request.formData();
        const rawPrompt = formData.get("prompt") as string || "Describe the image.";
        const prompt = `${rawPrompt}. Please include as much details as possible and answer in markdown format.`
        const imageType = formData.get('imgtype') as string;
        if (!imageType) {
            throw new Error('Image type not provided');
        }
        const file = formData.get("image") as File;
        if (!file) {
            throw new Error("No file uploaded");
        }
        if (file.size === 0) {
            throw new Error('File is empty');
        }

        // Convert file to buffer
        const buffer = Buffer.from(await file.arrayBuffer());
        if (buffer.length === 0) {
            throw new Error('File is empty');
        }

        // Convert file to base64
        const base64 = buffer.toString("base64");
        const blob = `data:${imageType};base64,${base64}`;
        console.log(`Base64 encoded image: ${blob.substring(0, 100)}...`);
        const payload = {
            id: crypto.randomUUID(),
            prompt: prompt,
            stream: true,
            blob
        };
        const body = JSON.stringify(payload);

        const API_ENDPOINT = process.env.API_ENDPOINT || "http://localhost:8080";

        // Send base64 to remote API (replace with your actual API endpoint)
        const response = await fetch(`${API_ENDPOINT}/extract-image-info`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body,
        });

        if (!response.ok) {
            console.log(`Response status: ${response.status}`);
            throw new Error("Failed to process image");
        }

        const reader = response.body?.getReader();
        if (!reader) {
            throw new Error("Failed to read response body");
        }

        const stream = new ReadableStream({
            async pull(controller) {
                const { done, value } = await reader.read();
                if (done) {
                    controller.close();
                    return;
                } else {
                    controller.enqueue(value);
                }
            }
        });

        return new Response(stream);
    } catch (error) {
        console.error("Processing failed:", error);
        return new Response(`Processing failed: ${error}`, { status: 500 });
    }
}
