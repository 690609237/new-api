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
import { ArrowUpRight, CheckCircle2, Download, FileCog } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

type GuideStep = {
  title: string
  description: string
  image?: string
  imageAlt?: string
}

export function Help() {
  const { t } = useTranslation()
  const steps: GuideStep[] = [
    {
      title: t('Download CC Switch'),
      description: t(
        'Download CC Switch to follow this example. You can skip this step if you prefer to edit the Codex configuration file another way.'
      ),
    },
    {
      title: t('Create an API key'),
      description: t(
        'Open the API Keys page on this site, create a key, and keep it ready for the next step.'
      ),
      image: '/help/ccswitch-step-api-key.png',
      imageAlt: t('API Keys page with the create button and key highlighted'),
    },
    {
      title: t('Open the key in CC Switch'),
      description: t(
        'Choose Open CC Switch for the API key, select Codex, and choose or enter a default model.'
      ),
      image: '/help/ccswitch-step-model.png',
      imageAlt: t(
        'CC Switch setup dialog with Codex and model fields highlighted'
      ),
    },
    {
      title: t('Import the configuration'),
      description: t(
        'Confirm the import. If the website address or API endpoint contains localhost, import it first and correct the address in the next step.'
      ),
      image: '/help/ccswitch-step-import.png',
      imageAlt: t('CC Switch import confirmation showing localhost addresses'),
    },
    {
      title: t('Replace the localhost addresses'),
      description: t(
        'Open configuration editing in CC Switch and replace both localhost addresses with this site address, including its port.'
      ),
      image: '/help/ccswitch-step-address.png',
      imageAlt: t(
        'CC Switch provider list with the imported configuration highlighted'
      ),
    },
    {
      title: t('Restart ChatGPT'),
      description: t(
        'Quit ChatGPT completely and open it again. The new Codex configuration is now ready to use.'
      ),
    },
  ]

  return (
    <PublicLayout showMainContainer={false}>
      <main className='bg-background min-h-[calc(100vh-4rem)]'>
        <section className='border-border/70 bg-muted/30 border-b'>
          <div className='mx-auto max-w-6xl px-4 py-12 sm:px-6 sm:py-16 lg:px-8'>
            <div className='max-w-4xl'>
              <p className='text-primary mb-3 text-sm font-semibold tracking-wide uppercase'>
                {t('Codex setup guide')}
              </p>
              <h1 className='text-foreground text-3xl font-semibold tracking-tight sm:text-4xl'>
                {t('Configure Codex to connect to this service')}
              </h1>
              <p className='text-muted-foreground mt-4 max-w-2xl text-base leading-7 sm:text-lg'>
                {t(
                  "Codex reads its API endpoint, credentials, and default model from a local configuration file. Connecting Codex to this service means updating that file with this site's API information."
                )}
              </p>
              <Alert
                role='note'
                className='bg-background/80 mt-7 max-w-3xl px-5 py-4 shadow-xs'
              >
                <FileCog
                  className='text-primary mt-0.5 size-5'
                  aria-hidden='true'
                />
                <AlertTitle className='text-base'>
                  {t('How it works')}
                </AlertTitle>
                <AlertDescription className='mt-1 leading-6'>
                  <p>
                    {t(
                      'You can edit the file manually or use a configuration manager. CC Switch is one example: it manages and updates the same Codex configuration file for you.'
                    )}
                  </p>
                </AlertDescription>
              </Alert>
            </div>
          </div>
        </section>

        <section
          className='mx-auto max-w-6xl px-4 py-10 sm:px-6 sm:py-14 lg:px-8'
          aria-labelledby='cc-switch-example-title'
        >
          <div className='mb-8 flex flex-col items-start justify-between gap-5 sm:mb-10 sm:flex-row sm:items-end'>
            <div className='max-w-3xl'>
              <p className='text-primary mb-2 text-sm font-semibold tracking-wide uppercase'>
                {t('Configuration example')}
              </p>
              <h2
                id='cc-switch-example-title'
                className='text-foreground text-2xl font-semibold tracking-tight sm:text-3xl'
              >
                {t('Example: Connect Codex with CC Switch')}
              </h2>
              <p className='text-muted-foreground mt-3 max-w-2xl text-sm leading-6 sm:text-base'>
                {t(
                  'The following steps use CC Switch as a convenient example. CC Switch is not required; any method that correctly updates the Codex configuration file will work.'
                )}
              </p>
            </div>
            <Button
              className='shrink-0'
              nativeButton={false}
              render={
                <a
                  href='https://ccswitch.io/zh/'
                  target='_blank'
                  rel='noopener noreferrer'
                />
              }
            >
              <Download aria-hidden='true' />
              {t('Download CC Switch')}
              <ArrowUpRight aria-hidden='true' />
            </Button>
          </div>

          <ol className='space-y-6 sm:space-y-8'>
            {steps.map((step, index) => (
              <li key={step.title}>
                <Card className='overflow-hidden py-0 shadow-sm'>
                  <CardContent className='p-0'>
                    <div
                      className={cn(
                        step.image &&
                          'grid lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)]'
                      )}
                    >
                      <div className='flex gap-4 p-6 sm:gap-5 sm:p-8'>
                        <div className='bg-primary text-primary-foreground flex size-9 shrink-0 items-center justify-center rounded-full text-sm font-semibold'>
                          {index + 1}
                        </div>
                        <div className='min-w-0 pt-0.5'>
                          <h3 className='text-foreground text-xl font-semibold'>
                            {step.title}
                          </h3>
                          <p className='text-muted-foreground mt-2 text-sm leading-6 sm:text-base'>
                            {step.description}
                          </p>
                          {index === 0 && (
                            <a
                              className='text-primary mt-4 inline-flex items-center gap-1 text-sm font-medium hover:underline'
                              href='https://ccswitch.io/zh/'
                              target='_blank'
                              rel='noopener noreferrer'
                            >
                              ccswitch.io/zh
                              <ArrowUpRight
                                className='size-4'
                                aria-hidden='true'
                              />
                            </a>
                          )}
                          {index === steps.length - 1 && (
                            <div className='text-primary mt-4 flex items-center gap-2 text-sm font-medium'>
                              <CheckCircle2
                                className='size-4'
                                aria-hidden='true'
                              />
                              {t('Setup complete')}
                            </div>
                          )}
                        </div>
                      </div>

                      {step.image && (
                        <div className='border-border bg-muted/40 border-t p-3 sm:p-4 lg:border-t-0 lg:border-l'>
                          <img
                            src={step.image}
                            alt={step.imageAlt}
                            loading='lazy'
                            className='border-border h-auto w-full rounded-lg border object-contain shadow-sm'
                          />
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              </li>
            ))}
          </ol>
        </section>
      </main>
    </PublicLayout>
  )
}
