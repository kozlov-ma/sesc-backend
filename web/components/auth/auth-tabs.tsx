"use client"

import { useState } from "react"
import { motion, AnimatePresence } from "framer-motion"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { LoginForm } from "@/components/auth/login-form"
import { AdminLoginForm } from "@/components/auth/admin-login-form"

export function AuthTabs() {
  const [activeTab, setActiveTab] = useState("user")

  return (
    <Tabs defaultValue="user" className="w-full" onValueChange={setActiveTab}>
      <TabsList className="grid w-full grid-cols-2">
        <TabsTrigger value="user">Пользователь</TabsTrigger>
        <TabsTrigger value="admin">Администратор</TabsTrigger>
      </TabsList>
      <AnimatePresence mode="wait">
        <motion.div
          key={activeTab}
          initial={{ opacity: 0, x: activeTab === "user" ? -20 : 20 }}
          animate={{ opacity: 1, x: 0 }}
          exit={{ opacity: 0, x: activeTab === "user" ? 20 : -20 }}
          transition={{ duration: 0.3 }}
        >
          <TabsContent value="user" className="mt-6">
            <LoginForm />
          </TabsContent>
          <TabsContent value="admin" className="mt-6">
            <AdminLoginForm />
          </TabsContent>
        </motion.div>
      </AnimatePresence>
    </Tabs>
  )
}
