"use client"

import axios from 'axios'
import { useBroadcast } from './hook/useBroadcast'
import { useState } from 'react'
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Button } from "@/components/ui/button"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select"

interface Payload {
  id: string
  timestamp: string
  data: any
}

interface Feedback {
  id?: string
  email: string
  description: string
  feedbackType: string
  createdAt: string
  updatedAt: string
}

const feedbackSchema = z.object({
  email: z.string().email({ message: "Invalid email address." }),
  description: z.string().min(5, {
    message: "Description must be at least 5 characters.",
  }),
  feedbackType: z.enum(["bug", "feature", "general"], {
    required_error: "Feedback type is required.",
  }),
})

type FeedbackFormValues = z.infer<typeof feedbackSchema>

function App() {
  const [feedbackList, setFeedbackList] = useState<Feedback[]>([])

  const form = useForm<FeedbackFormValues>({
    resolver: zodResolver(feedbackSchema),
    defaultValues: {
      email: "",
      description: "",
      feedbackType: "general",
    },
  })

  useBroadcast<Payload>("feedback.create", (data) => {
    console.log("Broadcast Create:", data)
  }, (error: Error) => {
    console.error("Broadcast Create Error:", error)
  })

  useBroadcast<Payload>("feedback.delete", (data) => {
    console.log("Broadcast Delete:", data)
  }, (error: Error) => {
    console.error("Broadcast Delete Error:", error)
  })

  useBroadcast<Payload>("feedback.update", (data) => {
    console.log("Broadcast Update:", data)
  }, (error: Error) => {
    console.error("Broadcast Update Error:", error)
  })

  const List = async () => {
    try {
      const res = await axios.get<Feedback[]>(`${import.meta.env.VITE_SERVER_URL}/feedback`)
      setFeedbackList(res.data)
    } catch (error) {
      console.error("List Error:", error)
    }
  }

  const Get = async (id: string) => {
    try {
      const res = await axios.get<Feedback>(`${import.meta.env.VITE_SERVER_URL}/feedback/${id}`)
      console.log("Get:", res.data)
    } catch (error) {
      console.error("Get Error:", error)
    }
  }

  const Create = async (data: FeedbackFormValues) => {
    try {
      const res = await axios.post<Feedback>(`${import.meta.env.VITE_SERVER_URL}/feedback`, data)
      console.log("Created:", res.data)
      List() // refresh list
    } catch (error) {
      console.error("Create Error:", error)
    }
  }

  const Update = async (id: string, data: Partial<Feedback>) => {
    try {
      const res = await axios.put<Feedback>(`${import.meta.env.VITE_SERVER_URL}/feedback/${id}`, data)
      console.log("Updated:", res.data)
      List()
    } catch (error) {
      console.error("Update Error:", error)
    }
  }

  const Delete = async (id: string) => {
    try {
      await axios.delete(`${import.meta.env.VITE_SERVER_URL}/feedback/${id}`)
      console.log("Deleted:", id)
      List()
    } catch (error) {
      console.error("Delete Error:", error)
    }
  }

  const onSubmit = (values: FeedbackFormValues) => {
    Create(values)
    form.reset()
  }

  return (
    <div className="p-6 max-w-xl mx-auto">
      <h2 className="text-2xl font-semibold mb-4">Submit Feedback</h2>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
          <FormField
            control={form.control}
            name="email"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Email</FormLabel>
                <FormControl><Input {...field} placeholder="you@example.com" /></FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="description"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Description</FormLabel>
                <FormControl><Textarea {...field} placeholder="Your feedback..." /></FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="feedbackType"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Feedback Type</FormLabel>
                <FormControl>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <SelectTrigger><SelectValue placeholder="Select type" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="bug">Bug</SelectItem>
                      <SelectItem value="feature">Feature</SelectItem>
                      <SelectItem value="general">General</SelectItem>
                    </SelectContent>
                  </Select>
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button type="submit">Submit</Button>
        </form>
      </Form>
    </div>
  )
}

export default App
