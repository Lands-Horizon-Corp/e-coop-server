"use client"

import axios from 'axios'
import { useEffect, useState } from 'react'
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Button } from "@/components/ui/button"
import {
  Form, FormControl, FormField, FormItem, FormLabel, FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import {
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
} from "@/components/ui/select"
import { useBroadcast } from '@/hook/useBroadcast'

interface Payload {
  id: string
  timestamp: string
  data: any
}

interface Feedback {
  id?: string
  email: string
  description: string
  feedback_type: string
  createdAt: string
  updatedAt: string
}

const feedbackSchema = z.object({
  email: z.string().email({ message: "Invalid email address." }),
  description: z.string().min(5, { message: "Description must be at least 5 characters." }),
  feedback_type: z.enum(["bug", "feature", "general"], {
    required_error: "Feedback type is required.",
  }),
})

type FeedbackFormValues = z.infer<typeof feedbackSchema>

function SampleForm() {
  const [feedbackList, setFeedbackList] = useState<Feedback[]>([])
  const [selectedFeedback, setSelectedFeedback] = useState<Feedback | null>(null)

  const form = useForm<FeedbackFormValues>({
    resolver: zodResolver(feedbackSchema),
    defaultValues: {
      email: "",
      description: "",
      feedback_type: "general",
    },
  })

  const fetchList = async () => {
    try {
      const res = await axios.get<Feedback[]>(`${import.meta.env.VITE_SERVER_URL}/feedback`, { withCredentials: true })
      setFeedbackList(res.data)
    } catch (error) {
      console.error("List Error:", error)
    }
  }

  const fetchFeedback = async (id: string) => {
    try {
      const res = await axios.get<Feedback>(`${import.meta.env.VITE_SERVER_URL}/feedback/${id}`, { withCredentials: true })
      setSelectedFeedback(res.data)
      form.reset({
        email: res.data.email,
        description: res.data.description,
        feedback_type: res.data.feedback_type as "bug" | "feature" | "general",
      })
    } catch (error) {
      console.error("Get Error:", error)
    }
  }

  const createFeedback = async (data: FeedbackFormValues) => {
    try {
      await axios.post(`${import.meta.env.VITE_SERVER_URL}/feedback`, data, { withCredentials: true })
      await fetchList()
      form.reset()
      setSelectedFeedback(null)
    } catch (error) {
      console.error("Create Error:", error)
    }
  }

  const updateFeedback = async (id: string, data: Partial<FeedbackFormValues>) => {
    try {
      await axios.put(`${import.meta.env.VITE_SERVER_URL}/feedback/${id}`, data, { withCredentials: true })
      await fetchList()
      form.reset()
      setSelectedFeedback(null)
    } catch (error) {
      console.error("Update Error:", error)
    }
  }

  const deleteFeedback = async (id: string) => {
    try {
      await axios.delete(`${import.meta.env.VITE_SERVER_URL}/feedback/${id}`, { withCredentials: true })
      setFeedbackList(prev => prev.filter(fb => fb.id !== id))
    } catch (error) {
      console.error("Delete Error:", error)
    }
  }

  const handleSubmit = (values: FeedbackFormValues) => {
    if (selectedFeedback?.id) {
      updateFeedback(selectedFeedback.id, values)
    } else {
      createFeedback(values)
    }
  }

  // useEffect(() => {
  //   fetchList()
  // }, [])

  useBroadcast<Feedback>("feedback.create", fetchList, console.error)
  useBroadcast<Payload>("feedback.update", fetchList, console.error)
  useBroadcast<Payload>("feedback.delete", fetchList, console.error)

  return (
    <div className="p-6 max-w-xxl mx-auto">
      <h2 className="text-2xl font-semibold mb-4">
        {selectedFeedback ? "Update Feedback" : "Submit Feedback"}
      </h2>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-5">
          {/* Fields */}
          <FormField control={form.control} name="email" render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl><Input {...field} placeholder="you@example.com" /></FormControl>
              <FormMessage />
            </FormItem>
          )} />
          <FormField control={form.control} name="description" render={({ field }) => (
            <FormItem>
              <FormLabel>Description</FormLabel>
              <FormControl><Textarea {...field} placeholder="Your feedback..." /></FormControl>
              <FormMessage />
            </FormItem>
          )} />
          <FormField control={form.control} name="feedback_type" render={({ field }) => (
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
          )} />
          <Button type="submit">{selectedFeedback ? "Update" : "Submit"}</Button>
        </form>
      </Form>

      {/* Feedback List */}
      <div className="mt-10">
        <h3 className="text-xl font-semibold mb-3">Feedback List</h3>
        <Button onClick={fetchList} className="mb-4">Refresh List</Button>
        <div className="overflow-x-auto">
          <table className="min-w-full table-auto border text-sm">
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
                  <td colSpan={5} className="text-center p-4 text-gray-500">No feedback yet.</td>
                </tr>
              ) : feedbackList.map((fb) => (
                <tr key={fb.id}>
                  <td className="border px-4 py-2">{fb.email}</td>
                  <td className="border px-4 py-2 capitalize">{fb.feedback_type}</td>
                  <td className="border px-4 py-2">{fb.description}</td>
                  <td className="border px-4 py-2">{new Date(fb.createdAt).toLocaleString()}</td>
                  <td className="border px-4 py-2 space-x-2">
                    <Button variant="outline" onClick={() => fetchFeedback(fb.id!)} size="sm">Edit</Button>
                    <Button variant="destructive" onClick={() => deleteFeedback(fb.id!)} size="sm">Delete</Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

export default SampleForm
