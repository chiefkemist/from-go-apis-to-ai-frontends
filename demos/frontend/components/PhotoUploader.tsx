'use client'

import { useState, useRef } from 'react'
import {flushSync, useFormStatus} from 'react-dom'
import ReactMarkdown from 'react-markdown'
import Form from 'next/form'
import Image from 'next/image'
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { extractImageInfo } from "@/actions/processImages"
import {Switch} from "@/components/ui/switch";
import {Label} from "@/components/ui/label";

export default function PhotoUploader() {
  const [results, setResults] = useState<string[]>([])
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [imageType, setImageType] = useState<string>('')
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [isStreaming, setIsStreaming] = useState(true)

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      const url = URL.createObjectURL(file)
      setPreviewUrl(url)
      setImageType(file.type) // Set the image type
    }
  }

  const handleSubmit = async (formData: FormData) => {
    setResults([])

    if (isStreaming) {
      const response = await fetch('/api/streaming/image-info', {
        method: 'POST',
        body: formData
      });

      if (!response.ok) {
        throw new Error('Upload failed')
      }

      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('Failed to get reader from response')
      }

      const decoder = new TextDecoder();

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split("\n\n");

        lines.forEach((line) => {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            if (data === '[DONE]') {
              reader.cancel();
            } else {
              if (data.startsWith('#') || data.startsWith('.') || data.startsWith('!') || data.startsWith('?')) {
                flushSync(() => {
                  setResults((results) => [...results, '<br />', data]);
                });
              } else {
                flushSync(() => {
                  setResults((results) => [...results, data]);
                });
              }
            }
          }
        });
      }
    } else {
      const response = await extractImageInfo(formData)

      if (response.success) {
        setResults([response.info]);
      } else {
        console.error(`Processing failed: ${response.error}`)
      }
    }
  }

  return (
    <Card className="w-[450px]">
      <CardHeader>
        <CardTitle>Upload Photo</CardTitle>
      </CardHeader>
      <CardContent>
        <Form action={handleSubmit}>
          <Input
              type="text"
              name="prompt"
              className="mb-4"
              placeholder="Prompt for details on the image..."
          />
          <input
              type="hidden"
              name="imgtype"
              value={imageType}
          />
          <Input
              type="file"
              accept="image/*"
              name="image"
              ref={fileInputRef}
              onChange={handleFileChange}
              className="mb-4"
          />
          <div className="flex items-center space-x-2 space-y-2">
            <Switch
                id="streaming-mode"
                checked={isStreaming}
                onCheckedChange={setIsStreaming}
            />
            <Label htmlFor="streaming-mode">Streaming mode</Label>
          </div>
          {previewUrl && (
              <>
                <div className="mb-4">
                  <Image src={previewUrl} alt="Preview" width={400} height={400} className="rounded-md"/>
                </div>
                <SubmitButton/>
              </>
          )}
        </Form>
      </CardContent>
      <CardFooter className="flex flex-col items-start">
        <h3 className="text-lg font-semibold mb-2">Results:</h3>
        <div className="w-full max-h-40 overflow-y-auto">
          {
            //results.map((result, index) => (
            //  <p key={index} className="text-sm">{result}</p>
            //))
            <div className="text-sm">
              {isStreaming &&
                  <ReactMarkdown>{results.join('')}</ReactMarkdown>
              }
              {!isStreaming &&
                  <ReactMarkdown>{results.join('\n')}</ReactMarkdown>
              }
            </div>
          }
        </div>
      </CardFooter>
    </Card>
  )
}

export function SubmitButton() {
  const status = useFormStatus()
  return (
    <Button type="submit" disabled={status.pending}>
      {status.pending ? 'Processing...' : 'Process Photo'}
    </Button>
  )
}
