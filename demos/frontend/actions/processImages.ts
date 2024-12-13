"use server";

import crypto from "node:crypto"

export async function extractImageInfo(formData: FormData) {
  try {
    const title = formData.get("title") as string;
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
    const blob = `data:image/webp;base64,${base64}`;
    console.log(`Base64 encoded image: ${blob.substring(0, 100)}...`);
    const payload = {
      id: crypto.randomUUID(),
      title: title,
      blob
    };
    const body = JSON.stringify(payload);

    // Send base64 to remote API (replace with your actual API endpoint)
    const response = await fetch("http://localhost:8080/extract-image-info", {
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

    // The response from the API should be handled by the server-sent events
    // So we don't need to do anything with it here

    return { success: true };
  } catch (error) {
    console.error("Processing failed:", error);
    return { success: false, error: (error as Error).message };
  }
}
