'use server'

export async function processPhoto(formData: FormData) {
  try {
    const file = formData.get('photo') as File
    if (!file) {
      throw new Error('No file uploaded')
    }

    // Convert file to base64
    const buffer = await file.arrayBuffer()
    const base64 = Buffer.from(buffer).toString('base64')

    // Send base64 to remote API (replace with your actual API endpoint)
    const response = await fetch('http://localhost:8080/process-photo', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ photo: base64 }),
    })

    if (!response.ok) {
      throw new Error('Failed to process photo')
    }

    // The response from the API should be handled by the server-sent events
    // So we don't need to do anything with it here

    return { success: true }
  } catch (error) {
    console.error('Processing failed:', error)
    return { success: false, error: (error as Error).message }
  }
}
