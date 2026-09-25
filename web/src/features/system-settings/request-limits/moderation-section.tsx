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
import { useMutation, useQuery } from '@tanstack/react-query'
import { CheckCircle2, ExternalLink, Loader2, XCircle } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import * as z from 'zod'

import { Button } from '@/components/ui/button'
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
import { Textarea } from '@/components/ui/textarea'
import { handleServerError } from '@/lib/handle-server-error'
import { requireServerSuccess } from '@/lib/server-error-message'

import {
  getSystemTask,
  startDailyReviewNow,
  testModerationEndpoint,
} from '../api'
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
    ModerationScoreThreshold: z.number().gt(0).max(1),
    ModerationAlertEmail: z
      .string()
      .refine(
        (value) => !value.trim() || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value),
        t('Enter a valid email or leave blank')
      ),
    ModerationAlertThreshold: z.number().int().min(1).max(1000000),
    ModerationCacheTTLSeconds: z.number().int().min(1).max(86400),
    ModerationExemptUserIDs: z.string(),
    ModerationExemptGroups: z.string(),
    ModerationSampleRate: z.number().int().min(0).max(100),
    ModerationForceUserIDs: z.string(),
    ModerationForceTokenIDs: z.string(),
    ModerationTimeoutSeconds: z.number().int().min(1).max(300),
    ModerationTimeoutWindowSeconds: z.number().int().min(1).max(86400),
    ModerationTimeoutThreshold: z.number().int().min(1).max(100),
    ModerationTimeoutPauseSeconds: z.number().int().min(1).max(86400),
    DailyReviewEnabled: z.boolean(),
    DailyReviewHour: z.number().int().min(0).max(23),
    DailyReviewPrompt: z.string().trim().min(1).max(20000),
    DailyReviewBaseURL: z.string().url(),
    DailyReviewModel: z.string().trim().min(1).max(128),
    DailyReviewAPIKey: z.string(),
  })

type ModerationFormValues = z.infer<ReturnType<typeof createModerationSchema>>

type ModerationSectionProps = {
  defaultValues: ModerationFormValues
}

export function ModerationSection({ defaultValues }: ModerationSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const [testState, setTestState] = useState<
    'idle' | 'testing' | 'success' | 'error'
  >('idle')
  const [testMessage, setTestMessage] = useState('')
  const [reviewTaskId, setReviewTaskId] = useState('')
  const runReview = useMutation({
    mutationFn: async () => requireServerSuccess(await startDailyReviewNow()),
    onSuccess: (response) => {
      if (response.data) setReviewTaskId(response.data.task_id)
    },
    onError: (error: Error) =>
      handleServerError(error, t('Failed to start daily review')),
  })
  const reviewTask = useQuery({
    queryKey: ['daily-review-task', reviewTaskId],
    queryFn: async () =>
      requireServerSuccess(await getSystemTask(reviewTaskId)),
    enabled: Boolean(reviewTaskId),
    refetchInterval: (query) => {
      const status = query.state.data?.data?.status
      return status === 'succeeded' || status === 'failed' ? false : 3000
    },
  })
  const schema = createModerationSchema(t)
  const form = useForm<ModerationFormValues>({
    resolver: zodResolver(schema),
    mode: 'onChange',
    defaultValues,
  })

  useEffect(() => {
    form.reset(defaultValues)
  }, [defaultValues, form])

  const handleTestConnection = async () => {
    const values = form.getValues()
    if (!values.ModerationBaseURL.trim() || !values.ModerationAPIKey.trim()) {
      setTestState('error')
      setTestMessage(t('Enter a moderation base URL and API key first.'))
      return
    }

    setTestState('testing')
    setTestMessage('')
    try {
      const response = await testModerationEndpoint({
        base_url: values.ModerationBaseURL,
        api_key: values.ModerationAPIKey,
        model: values.ModerationModel,
      })
      if (!response.success) {
        throw new Error(
          response.message || t('Moderation connection test failed.')
        )
      }
      setTestState('success')
      setTestMessage(
        response.data?.flagged
          ? t('Connection succeeded; the test text was flagged.')
          : t('Connection succeeded; the test text was allowed.')
      )
    } catch (error) {
      setTestState('error')
      setTestMessage(
        error instanceof Error
          ? error.message
          : t('Moderation connection test failed.')
      )
    }
  }

  const onSubmit = async (values: ModerationFormValues) => {
    const updates = Object.entries(values).filter(([key, value]) => {
      if (
        (key === 'ModerationAPIKey' || key === 'DailyReviewAPIKey') &&
        value === ''
      ) {
        return false
      }
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
            <div className='space-y-3' data-moderation-layout='connection'>
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
              <div className='flex flex-wrap items-center gap-3 rounded-lg border border-dashed p-3'>
                <Button
                  type='button'
                  variant='outline'
                  onClick={() => void handleTestConnection()}
                  disabled={testState === 'testing'}
                >
                  {testState === 'testing' ? (
                    <Loader2 className='animate-spin' />
                  ) : (
                    <CheckCircle2 />
                  )}
                  {t('Test moderation connection')}
                </Button>
                <a
                  href='https://platform.openai.com/usage'
                  target='_blank'
                  rel='noreferrer'
                  className='text-primary inline-flex items-center gap-1 text-sm underline-offset-4 hover:underline'
                >
                  {t('View OpenAI usage statistics')}
                  <ExternalLink className='size-3' aria-hidden='true' />
                </a>
                {testState !== 'idle' && testMessage && (
                  <div
                    className={`inline-flex items-center gap-1 text-sm ${
                      testState === 'success'
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-destructive'
                    }`}
                    role='status'
                  >
                    {testState === 'success' ? (
                      <CheckCircle2 className='size-4' aria-hidden='true' />
                    ) : (
                      <XCircle className='size-4' aria-hidden='true' />
                    )}
                    <span>{testMessage}</span>
                  </div>
                )}
              </div>
            </div>
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
              name='ModerationScoreThreshold'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Moderation score threshold')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={0}
                      max={1}
                      step='any'
                      {...field}
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Reject when any category score reaches this value (greater than 0 and at most 1). Default: 0.6.'
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
                    {t(
                      'Seconds to reuse sensitive-word and moderation results for identical user content.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='ModerationSampleRate'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Moderation sample rate')}</FormLabel>
                  <FormControl>
                    <div className='flex items-center gap-2'>
                      <Input
                        type='number'
                        min={0}
                        max={100}
                        step={1}
                        {...field}
                        onChange={(e) =>
                          field.onChange(Number.parseInt(e.target.value) || 0)
                        }
                      />
                      <span className='text-muted-foreground text-sm'>%</span>
                    </div>
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Percentage of non-exempt users selected for moderation. 100% checks everyone.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className='text-muted-foreground rounded-lg border border-dashed p-3 text-sm'>
            {t(
              'Timeout protection pauses moderation temporarily when the upstream service is unstable.'
            )}
          </div>

          <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-4'>
            {(
              [
                [
                  'ModerationTimeoutSeconds',
                  'Moderation request timeout',
                  1,
                  300,
                  'Seconds before a moderation request is skipped.',
                ],
                [
                  'ModerationTimeoutWindowSeconds',
                  'Timeout judgment window',
                  1,
                  86400,
                  'Window used to count consecutive timeouts.',
                ],
                [
                  'ModerationTimeoutThreshold',
                  'Consecutive timeout threshold',
                  1,
                  100,
                  'Open the pause after this many timeouts.',
                ],
                [
                  'ModerationTimeoutPauseSeconds',
                  'Moderation pause duration',
                  1,
                  86400,
                  'Seconds to pause moderation after the threshold.',
                ],
              ] as const
            ).map(([name, label, min, max, description]) => (
              <FormField
                key={name}
                control={form.control}
                name={name}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t(label)}</FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        min={min}
                        max={max}
                        step={1}
                        {...field}
                        onChange={(e) =>
                          field.onChange(Number.parseInt(e.target.value) || min)
                        }
                      />
                    </FormControl>
                    <FormDescription>{t(description)}</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            ))}
          </div>

          <div
            className='grid gap-4 md:grid-cols-2'
            data-moderation-layout='audience-rules'
          >
            <FormField
              control={form.control}
              name='ModerationExemptUserIDs'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Exempt user IDs')}</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={5}
                      placeholder={t('One user ID per line')}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'These users bypass moderation. Commas and line breaks are supported.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='ModerationExemptGroups'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Exempt user groups')}</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={5}
                      placeholder={t('One group per line')}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Users in these groups bypass moderation. Matching is case-insensitive.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='ModerationForceUserIDs'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Required moderation user IDs')}</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={5}
                      placeholder={t('One user ID per line')}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'These users are always moderated, even when they or their group are exempt.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='ModerationForceTokenIDs'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Required moderation token IDs')}</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={5}
                      placeholder={t('One token ID per line')}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'These tokens are always moderated, even when their user or group is exempt.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className='space-y-4 rounded-lg border p-4'>
            <div>
              <h3 className='font-semibold'>{t('Daily content review')}</h3>
              <p className='text-muted-foreground text-sm'>
                {t(
                  'Review yesterday’s user-message logs at the configured server-local hour. The manual run reviews today.'
                )}
              </p>
            </div>
            <FormField
              control={form.control}
              name='DailyReviewEnabled'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>{t('Enable daily review')}</FormLabel>
                    <FormDescription>
                      {t(
                        'Send user-message logs to the configured Responses API for inspection.'
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
            <div className='grid gap-4 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='DailyReviewHour'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Daily review hour')}</FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        min={0}
                        max={23}
                        step={1}
                        {...field}
                        onChange={(event) =>
                          field.onChange(Number(event.target.value))
                        }
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Server-local hour, 0–23.')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='DailyReviewModel'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Daily review model')}</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='DailyReviewBaseURL'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Daily review API base URL')}</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormDescription>
                      {t('Include /v1; localhost may use HTTP.')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='DailyReviewAPIKey'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Daily review API key')}</FormLabel>
                    <FormControl>
                      <Input
                        type='password'
                        autoComplete='new-password'
                        placeholder={t('Leave blank to keep the existing key')}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
            <FormField
              control={form.control}
              name='DailyReviewPrompt'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Daily review prompt')}</FormLabel>
                  <FormControl>
                    <Textarea rows={10} {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className='flex flex-wrap items-center gap-3'>
              <Button
                type='button'
                variant='outline'
                disabled={
                  runReview.isPending ||
                  reviewTask.data?.data?.status === 'running' ||
                  reviewTask.data?.data?.status === 'pending'
                }
                onClick={() => runReview.mutate()}
              >
                {runReview.isPending ? (
                  <Loader2 className='animate-spin' />
                ) : null}
                {t('Review today now')}
              </Button>
              {reviewTask.data?.data ? (
                <span role='status' className='text-muted-foreground text-sm'>
                  {t('Review status')}: {t(reviewTask.data.data.status)}
                  {reviewTask.data.data.error
                    ? ` — ${reviewTask.data.data.error}`
                    : ''}
                </span>
              ) : null}
            </div>
          </div>
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
