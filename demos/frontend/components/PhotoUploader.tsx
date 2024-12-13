'use client'

import { useState, useRef } from 'react'
import Image from 'next/image'
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { extractImageInfo } from "@/actions/processImages"

export default function PhotoUploader() {
  const [processing, setProcessing] = useState(false)
  const [results, setResults] = useState<string[]>([])
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      const url = URL.createObjectURL(file)
      setPreviewUrl(url)
    }
  }

  const handleSubmit = async (formData: FormData) => {
    setProcessing(true)
    setResults([])

    const response = await extractImageInfo(formData)

    if (response.success) {
      const eventSource = new EventSource('/api/process-events')

      eventSource.onmessage = (event) => {
        setResults((prevResults) => [...prevResults, event.data])
      }

      eventSource.onerror = (error) => {
        console.error(`EventSource failed: ${error}`)
        eventSource.close()
        setProcessing(false)
      }

      eventSource.addEventListener('done', () => {
        eventSource.close()
        setProcessing(false)
      })
    } else {
      console.error(`Processing failed: ${response.error}`)
      setProcessing(false)
    }
  }

  return (
    <Card className="w-[350px]">
      <CardHeader>
        <CardTitle>Upload Photo</CardTitle>
      </CardHeader>
      <CardContent>
        <form action={handleSubmit}>
          <Input
            type="text"
            name="title"
            className="mb-4"
          />
          <Input
            type="file"
            accept="image/*"
            name="image"
            ref={fileInputRef}
            onChange={handleFileChange}
            className="mb-4"
          />
          {previewUrl && (
            <div className="mb-4">
              <Image src={previewUrl} alt="Preview" width={300} height={300} className="rounded-md" />
            </div>
          )}
          <Button type="submit" disabled={processing || !previewUrl}>
            {processing ? 'Processing...' : 'Process Photo'}
          </Button>
        </form>
      </CardContent>
      <CardFooter className="flex flex-col items-start">
        <h3 className="text-lg font-semibold mb-2">Results:</h3>
        <div className="w-full max-h-40 overflow-y-auto">
          {results.map((result, index) => (
            <p key={index} className="text-sm">{result}</p>
          ))}
        </div>
      </CardFooter>
    </Card>
  )
}
