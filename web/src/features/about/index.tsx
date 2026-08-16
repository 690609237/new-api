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
import { Mail, MessageCircle, Smartphone, Users } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

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
          <div className='relative mx-auto max-w-6xl px-4 py-12 sm:px-6 sm:py-16 lg:px-8'>
            <div className='max-w-3xl'>
              <p className='text-primary mb-3 text-sm font-semibold tracking-wide uppercase'>
                {t('About the author')}
              </p>
              <h1 className='text-foreground text-3xl font-semibold tracking-tight sm:text-4xl'>
                {t('Get in touch')}
              </h1>
              <p className='text-muted-foreground mt-4 max-w-2xl text-base leading-7 sm:text-lg'>
                {t(
                  'For product questions, usage feedback, or collaboration, contact the author through any channel below.'
                )}
              </p>
            </div>
          </div>
        </section>

        <section
          className='mx-auto max-w-6xl px-4 py-10 sm:px-6 sm:py-14 lg:px-8'
          aria-labelledby='author-contacts-title'
        >
          <h2
            id='author-contacts-title'
            className='text-foreground mb-6 text-2xl font-semibold tracking-tight'
          >
            {t('Author contacts')}
          </h2>

          <address className='not-italic'>
            <ul className='grid gap-4 sm:grid-cols-2'>
              {contacts.map((contact) => {
                const Icon = contact.icon

                return (
                  <li key={contact.label}>
                    <Card className='h-full shadow-sm'>
                      <CardHeader className='grid grid-cols-[auto_1fr] items-center gap-x-3'>
                        <div className='bg-primary/10 text-primary flex size-10 items-center justify-center rounded-lg'>
                          <Icon className='size-5' aria-hidden='true' />
                        </div>
                        <CardTitle className='flex items-center gap-2'>
                          {t(contact.label)}
                          {'recommended' in contact && (
                            <Badge variant='secondary'>{t('Recommended')}</Badge>
                          )}
                        </CardTitle>
                      </CardHeader>
                      <CardContent className='space-y-2'>
                        {'href' in contact ? (
                          <a
                            className='text-foreground text-lg font-medium underline-offset-4 hover:underline'
                            href={contact.href}
                          >
                            {contact.value}
                          </a>
                        ) : (
                          <p className='text-foreground text-lg font-medium'>
                            {contact.value}
                          </p>
                        )}
                        {'description' in contact && (
                          <p className='text-muted-foreground text-sm leading-6'>
                            {t(contact.description)}
                          </p>
                        )}
                      </CardContent>
                    </Card>
                  </li>
                )
              })}
            </ul>
          </address>

        </section>
      </main>
    </PublicLayout>
  )
}
