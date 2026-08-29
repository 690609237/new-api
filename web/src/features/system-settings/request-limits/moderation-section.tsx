/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import * as z from 'zod'

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'

import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

const createModerationSchema = (t: (key: string) => string) =>
  z.object({
    ModerationEnabled: z.boolean(),
    ModerationBeforeChannel: z.boolean(),
    ModerationBaseURL: z.string(),
    ModerationAPIKey: z.string(),
    ModerationModel: z.string(),
    ModerationAlertEmail: z
      .string()
      .refine(
        (value) => !value.trim() || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value),
        t('Enter a valid email or leave blank')
      ),
    ModerationAlertThreshold: z.number().int().min(1).max(1000000),
    ModerationCacheTTLSeconds: z.number().int().min(1).max(86400),
  })

type ModerationFormValues = z.infer<ReturnType<typeof createModerationSchema>>

type ModerationSectionProps = {
  defaultValues: ModerationFormValues
}

export function ModerationSection({ defaultValues }: ModerationSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const schema = createModerationSchema(t)
  const form = useForm<ModerationFormValues>({
    resolver: zodResolver(schema),
    mode: 'onChange',
    defaultValues,
  })

  useEffect(() => {
    form.reset(defaultValues)
  }, [defaultValues, form])

  const onSubmit = async (values: ModerationFormValues) => {
    const updates = Object.entries(values).filter(([key, value]) => {
      if (key === 'ModerationAPIKey' && value === '') return false
      return value !== defaultValues[key as keyof ModerationFormValues]
    })

    for (const [key, value] of updates) {
      await updateOption.mutateAsync({ key, value: value ?? '' })
    }
  }

  return (
    <SettingsSection title={t('Content Moderation')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
            saveLabel='Save moderation settings'
          />

          <div className='space-y-4'>
            <FormField
              control={form.control}
              name='ModerationEnabled'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>{t('Enable content moderation')}</FormLabel>
                    <FormDescription>
                      {t(
                        'Use an OpenAI-compatible moderation endpoint to scan user prompts.'
                      )}
                    </FormDescription>
                  </SettingsSwitchContent>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </SettingsSwitchItem>
              )}
            />

            <FormField
              control={form.control}
              name='ModerationBeforeChannel'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>
                      {t('Moderate before channel selection')}
                    </FormLabel>
                    <FormDescription>
                      {t(
                        'Scan prompts before selecting an upstream channel. This may add latency.'
                      )}
                    </FormDescription>
                  </SettingsSwitchContent>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </SettingsSwitchItem>
              )}
            />
          </div>

          <div className='grid gap-4 md:grid-cols-2'>
            <FormField
              control={form.control}
              name='ModerationBaseURL'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Moderation base URL')}</FormLabel>
                  <FormControl>
                    <Input placeholder='https://api.openai.com/v1' {...field} />
                  </FormControl>
                  <FormDescription>
                    {t('The endpoint should expose POST /moderations.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='ModerationModel'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Moderation model')}</FormLabel>
                  <FormControl>
                    <Input placeholder='omni-moderation-latest' {...field} />
                  </FormControl>
                  <FormDescription>
                    {t('Defaults to omni-moderation-latest when blank.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='ModerationAPIKey'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Moderation API key')}</FormLabel>
                  <FormControl>
                    <Input
                      type='password'
                      autoComplete='new-password'
                      placeholder={t('Leave blank to keep the existing key')}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'The key is write-only and is never shown after saving.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='ModerationAlertEmail'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Moderation alert email')}</FormLabel>
                  <FormControl>
                    <Input
                      type='email'
                      placeholder={t('Optional alert recipient')}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Receive an email after repeated moderation upstream failures.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='ModerationAlertThreshold'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Moderation alert threshold')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={1}
                      max={1000000}
                      step={1}
                      {...field}
                      onChange={(e) =>
                        field.onChange(Number.parseInt(e.target.value) || 1)
                      }
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Failures within 30 minutes before an alert is sent.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='ModerationCacheTTLSeconds'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Moderation cache TTL')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={1}
                      max={86400}
                      step={1}
                      {...field}
                      onChange={(e) =>
                        field.onChange(Number.parseInt(e.target.value) || 1)
                      }
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Seconds to reuse a successful moderation result.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
