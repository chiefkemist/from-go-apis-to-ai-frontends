"use server";

import crypto from "node:crypto"

export async function extractImageInfo(formData: FormData) {
  try {
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

    const json = await response.json();

    console.log(JSON.stringify(json, null, 4));

    // The response from the API should be handled by the server-sent events
    // So we don't need to do anything with it here

    return { success: true, ...json };
  } catch (error) {
    console.error("Processing failed:", error);
    return { success: false, error: (error as Error).message };
  }
}
