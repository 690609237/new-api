/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
import {
  Loader2,
  Mail,
  MessageCircle,
  Send,
  Smartphone,
  Users,
} from 'lucide-react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { z } from 'zod'

import { PublicLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { handleServerError } from '@/lib/handle-server-error'
import { createServerError } from '@/lib/server-error-message'
import { useAuthStore } from '@/stores/auth-store'

import { sendContactEmail } from './api'

const contacts = [
  { label: 'QQ', value: '1549277597', icon: MessageCircle },
  { label: 'WeChat', value: 'ModelPass', icon: Smartphone },
  {
    label: 'Email',
    value: '1549277597@qq.com',
    href: 'mailto:1549277597@qq.com',
    icon: Mail,
  },
  {
    label: 'QQ Group',
    value: '450997742',
    description: 'Occasional AI application exchange activities are held here.',
    recommended: true,
    icon: Users,
  },
] as const

const contactEmailSchema = z.object({
  subject: z
    .string()
    .trim()
    .min(1, 'Please enter an email subject')
    .max(120, 'The email subject must be 120 characters or fewer'),
  content: z
    .string()
    .trim()
    .min(1, 'Please enter email content')
    .max(5000, 'The email content must be 5000 characters or fewer'),
})

type ContactEmailFormValues = z.infer<typeof contactEmailSchema>

function ContactMethods() {
  const { t } = useTranslation()

  return (
    <address className='not-italic'>
      <ul
        className='grid gap-2 sm:grid-cols-2 lg:grid-cols-4'
        aria-label={t('Author contacts')}
      >
        {contacts.map((contact) => {
          const Icon = contact.icon
          return (
            <li key={contact.label}>
              <div className='border-border/70 bg-card h-full rounded-xl border px-3 py-2 shadow-sm'>
                <div className='flex items-center gap-2'>
                  <div className='bg-primary/10 text-primary flex size-7 items-center justify-center rounded-md'>
                    <Icon className='size-3.5' aria-hidden='true' />
                  </div>
                  <div className='text-foreground flex items-center gap-1.5 text-xs font-semibold'>
                    {t(contact.label)}
                    {'recommended' in contact && (
                      <Badge variant='secondary'>{t('Recommended')}</Badge>
                    )}
                  </div>
                </div>
                <div className='mt-1.5 pl-9'>
                  {'href' in contact ? (
                    <a
                      className='text-foreground text-sm font-medium underline-offset-4 hover:underline'
                      href={contact.href}
                    >
                      {contact.value}
                    </a>
                  ) : (
                    <p className='text-foreground text-sm font-medium'>
                      {contact.value}
                    </p>
                  )}
                  {'description' in contact && (
                    <p className='text-muted-foreground text-[11px] leading-4'>
                      {t(contact.description)}
                    </p>
                  )}
                </div>
              </div>
            </li>
          )
        })}
      </ul>
    </address>
  )
}

function ContactEmailForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const user = useAuthStore((state) => state.auth.user)
  const form = useForm<ContactEmailFormValues>({
    resolver: zodResolver(contactEmailSchema),
    defaultValues: { subject: '', content: '' },
  })
  const isSubmitting = form.formState.isSubmitting

  const redirectToSignIn = async () => {
    await navigate({ to: '/sign-in', search: { redirect: '/about' } })
  }

  const onSubmit = async (values: ContactEmailFormValues) => {
    if (!user) {
      await redirectToSignIn()
      return
    }
    try {
      const response = await sendContactEmail(values)
      if (!response.success) {
        throw createServerError(response, t('Failed to send email'))
      }
      form.reset()
      toast.success(t('Email sent successfully'))
    } catch (error) {
      handleServerError(error, t('Failed to send email'))
    }
  }

  return (
    <div className='border-border/70 bg-card rounded-xl border p-4 shadow-sm sm:p-5'>
      <div className='mb-4 flex items-start justify-between gap-4'>
        <div>
          <h2 className='text-foreground flex flex-wrap items-baseline gap-x-2 gap-y-1 text-lg font-semibold tracking-tight'>
            <span>{t('Leave me a message')}</span>
            <span className='text-muted-foreground text-xs font-normal'>
              {t('Signed-in users can send up to 3 messages per day.')}
            </span>
          </h2>
        </div>
        <Mail className='text-primary mt-1 size-5' aria-hidden='true' />
      </div>
      <Form {...form}>
        <form className='space-y-4' onSubmit={form.handleSubmit(onSubmit)}>
          <FormField
            control={form.control}
            name='subject'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Email subject')}</FormLabel>
                <FormControl>
                  <Input
                    {...field}
                    maxLength={120}
                    placeholder={t('Enter an email subject')}
                    disabled={isSubmitting}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='content'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Message')}</FormLabel>
                <FormControl>
                  <Textarea
                    {...field}
                    maxLength={5000}
                    rows={8}
                    placeholder={t(
                      'Describe your question, feedback, or collaboration idea'
                    )}
                    disabled={isSubmitting}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button
            type={user ? 'submit' : 'button'}
            disabled={isSubmitting}
            onClick={user ? undefined : redirectToSignIn}
          >
            {isSubmitting ? (
              <Loader2 className='animate-spin' aria-hidden='true' />
            ) : (
              <Send aria-hidden='true' />
            )}
            {user ? t('Send email') : t('Sign in to send')}
          </Button>
        </form>
      </Form>
    </div>
  )
}

function UsageAgreement() {
  const { t } = useTranslation()
  const rules = [
    'Use AI model APIs responsibly and do not engage in illegal or prohibited activities.',
    'Ensure your requests respect applicable laws, provider terms, and the privacy rights of others.',
    'Do not use the service to bypass provider limits, conduct attacks, or generate large-scale abusive traffic.',
    'Keep API keys and account credentials private; sharing or reselling access is not allowed.',
    'During routine inspections, high-risk violations may trigger an email reminder. Repeated violations may result in account suspension.',
    'Please follow site announcements. AI model API pricing multipliers are affected by market conditions and may change frequently.',
    'List prices in model pricing are shown in the original currency: foreign models use foreign-currency units, while domestic models use CNY, aligned with official pricing.',
    'Usage records and settlement results are based on platform logs; contact support promptly if you find a discrepancy.',
  ]
  return (
    <div className='border-border/70 bg-card rounded-xl border p-4 shadow-sm sm:p-5'>
      <h2 className='text-foreground text-lg font-semibold tracking-tight'>
        {t('ModelPass usage agreement')}
      </h2>
      <p className='text-muted-foreground mt-2 text-sm leading-6'>
        {t(
          'These rules help keep the service safe, stable, and available to everyone.'
        )}
      </p>
      <ol className='text-muted-foreground mt-4 list-decimal space-y-2 pl-5 text-sm leading-6'>
        {rules.map((rule) => (
          <li key={rule}>{t(rule)}</li>
        ))}
      </ol>
    </div>
  )
}

export function About() {
  const { t } = useTranslation()
  return (
    <PublicLayout showMainContainer={false}>
      <main className='bg-background min-h-[calc(100vh-4rem)]'>
        <section className='border-border/70 bg-muted/30 relative overflow-hidden border-b'>
          <div
            className='bg-primary/10 pointer-events-none absolute -top-28 right-[8%] size-72 rounded-full blur-3xl'
            aria-hidden='true'
          />
          <div className='relative mx-auto max-w-6xl px-4 py-10 sm:px-6 sm:py-14 lg:px-8'>
            <h1 className='text-foreground text-3xl font-semibold tracking-tight sm:text-4xl'>
              {t('About ModelPass')}
            </h1>
          </div>
        </section>
        <section className='mx-auto max-w-6xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8'>
          <Tabs defaultValue='contact'>
            <TabsList className='w-full sm:w-auto'>
              <TabsTrigger value='contact'>{t('Contact us')}</TabsTrigger>
              <TabsTrigger value='agreement'>
                {t('Usage agreement')}
              </TabsTrigger>
            </TabsList>
            <TabsContent value='contact' className='mt-6 space-y-6'>
              <ContactMethods />
              <ContactEmailForm />
            </TabsContent>
            <TabsContent value='agreement' className='mt-6'>
              <UsageAgreement />
            </TabsContent>
          </Tabs>
        </section>
      </main>
    </PublicLayout>
  )
}
