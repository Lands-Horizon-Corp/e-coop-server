"use client"

import axios from 'axios'
import { useBroadcast } from './hook/useBroadcast'
import { useEffect, useState } from 'react'
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

  useEffect(()=>{
    List()
    return () => {}
  }, [])

  useBroadcast<Feedback>("feedback.create", (data) => {
    console.log(data)
  }, (error: Error) => {
    console.error("Broadcast Create Error:", error)
  })

  useBroadcast<Payload>("feedback.delete", (data) => {
    console.log(data)
  }, (error: Error) => {
    console.error("Broadcast Delete Error:", error)
  })

  useBroadcast<Payload>("feedback.update", (data) => {
    console.log(data)
  }, (error: Error) => {
    console.error("Broadcast Update Error:", error)
  })
  
  const List = async () => {
    try {
      const res = await axios.get<Feedback[]>(
        `${import.meta.env.VITE_SERVER_URL}/feedback`,
        { withCredentials: true }
      )
      setFeedbackList(res.data)
      console.log("fetch")
    } catch (error) {
      console.error("List Error:", error)
    }
  }

  const Get = async (id: string) => {
    try {
      const res = await axios.get<Feedback>(
        `${import.meta.env.VITE_SERVER_URL}/feedback/${id}`,
        { withCredentials: true }
      )
      console.log("Get:", res.data)
    } catch (error) {
      console.error("Get Error:", error)
    }
  }

  const Create = async (data: FeedbackFormValues) => {
    try {
      const res = await axios.post<Feedback>(
        `${import.meta.env.VITE_SERVER_URL}/feedback`,
        data,
        { withCredentials: true }
      )
      console.log("Created:", res.data)
     
    } catch (error) {
      console.error("Create Error:", error)
    }
  }

  const Update = async (id: string, data: Partial<Feedback>) => {
    try {
      const res = await axios.put<Feedback>(
        `${import.meta.env.VITE_SERVER_URL}/feedback/${id}`,
        data,
        { withCredentials: true }
      )
      console.log("Updated:", res.data)
    } catch (error) {
      console.error("Update Error:", error)
    }
  }

  const Delete = async (id: string) => {
    try {
      await axios.delete(
        `${import.meta.env.VITE_SERVER_URL}/feedback/${id}`,
        { withCredentials: true }
      )
      console.log("Deleted:", id)
    } catch (error) {
      console.error("Delete Error:", error)
    }
  }


  const onSubmit = (values: FeedbackFormValues) => {
    Create(values)
    form.reset()
  }

  return (
    <div className="p-6 max-w-xxl mx-auto">
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
      <div className="mt-10">
        <h3 className="text-xl font-semibold mb-3">Feedback List</h3>
        <Button onClick={List} className="mb-4">Refresh List</Button>
        <div className="overflow-x-auto">
          <table className="min-w-full table-auto border border-gray-200 text-sm">
            <thead className="bg-gray-100">
              <tr>
                <th className="border px-4 py-2 text-left">Email</th>
                <th className="border px-4 py-2 text-left">Type</th>
                <th className="border px-4 py-2 text-left">Description</th>
                <th className="border px-4 py-2 text-left">Created At</th>
                <th className="border px-4 py-2 text-left">Actions</th>
              </tr>
            </thead>
            <tbody>
              {feedbackList.length === 0 ? (
                <tr>
                  <td colSpan={5} className="text-center p-4 text-gray-500">
                    No feedback yet.
                  </td>
                </tr>
              ) : (
                feedbackList.map((fb) => (
                  <tr key={fb.id}>
                    <td className="border px-4 py-2">{fb.email}</td>
                    <td className="border px-4 py-2 capitalize">{fb.feedbackType}</td>
                    <td className="border px-4 py-2">{fb.description}</td>
                    <td className="border px-4 py-2">{new Date(fb.createdAt).toLocaleString()}</td>
                    <td className="border px-4 py-2 space-x-2">
                      <Button
                        variant="outline"
                        onClick={() => Get(fb.id!)}
                        size="sm"
                      >
                        View
                      </Button>
                      <Button
                        variant="destructive"
                        onClick={() => Delete(fb.id!)}
                        size="sm"
                      >
                        Delete
                      </Button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

    </div>
  )
}

export default App
