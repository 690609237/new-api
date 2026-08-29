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
import fs from 'node:fs/promises'
import path from 'node:path'

const LOCALES_DIR = path.resolve('src/i18n/locales')

function stableStringify(obj) {
  return `${JSON.stringify(obj, null, 2)}\n`
}

const newKeys = {
  en: {
    'Content Moderation': 'Content Moderation',
    'Save moderation settings': 'Save moderation settings',
    'Enable content moderation': 'Enable content moderation',
    'Use an OpenAI-compatible moderation endpoint to scan user prompts.':
      'Use an OpenAI-compatible moderation endpoint to scan user prompts.',
    'Moderate before channel selection': 'Moderate before channel selection',
    'Scan prompts before selecting an upstream channel. This may add latency.':
      'Scan prompts before selecting an upstream channel. This may add latency.',
    'Moderation base URL': 'Moderation base URL',
    'The endpoint should expose POST /moderations.':
      'The endpoint should expose POST /moderations.',
    'Moderation model': 'Moderation model',
    'Defaults to omni-moderation-latest when blank.':
      'Defaults to omni-moderation-latest when blank.',
    'Moderation API key': 'Moderation API key',
    'The key is write-only and is never shown after saving.':
      'The key is write-only and is never shown after saving.',
    'Moderation alert email': 'Moderation alert email',
    'Optional alert recipient': 'Optional alert recipient',
    'Receive an email after repeated moderation upstream failures.':
      'Receive an email after repeated moderation upstream failures.',
    'Moderation alert threshold': 'Moderation alert threshold',
    'Failures within 30 minutes before an alert is sent.':
      'Failures within 30 minutes before an alert is sent.',
    'Moderation cache TTL': 'Moderation cache TTL',
    'Seconds to reuse a successful moderation result.':
      'Seconds to reuse a successful moderation result.',
  },
  zh: {
    'Content Moderation': '内容审核',
    'Save moderation settings': '保存审核设置',
    'Enable content moderation': '启用内容审核',
    'Use an OpenAI-compatible moderation endpoint to scan user prompts.':
      '使用兼容 OpenAI 的审核接口扫描用户提示词。',
    'Moderate before channel selection': '在选择渠道前审核',
    'Scan prompts before selecting an upstream channel. This may add latency.':
      '在选择上游渠道前扫描提示词，可能会增加延迟。',
    'Moderation base URL': '审核接口地址',
    'The endpoint should expose POST /moderations.':
      '接口应提供 POST /moderations 路径。',
    'Moderation model': '审核模型',
    'Defaults to omni-moderation-latest when blank.':
      '留空时使用 omni-moderation-latest。',
    'Moderation API key': '审核 API 密钥',
    'The key is write-only and is never shown after saving.':
      '密钥仅支持写入，保存后不会显示。',
    'Moderation alert email': '审核告警邮箱',
    'Optional alert recipient': '可选的告警接收邮箱',
    'Receive an email after repeated moderation upstream failures.':
      '审核上游连续失败后接收邮件通知。',
    'Moderation alert threshold': '审核告警阈值',
    'Failures within 30 minutes before an alert is sent.':
      '30 分钟内达到此失败次数后发送告警。',
    'Moderation cache TTL': '审核缓存时长',
    'Seconds to reuse a successful moderation result.':
      '成功审核结果的复用时长（秒）。',
  },
  'zh-TW': {
    'Content Moderation': '內容審核',
    'Save moderation settings': '儲存審核設定',
    'Enable content moderation': '啟用內容審核',
    'Use an OpenAI-compatible moderation endpoint to scan user prompts.':
      '使用相容 OpenAI 的審核端點掃描使用者提示詞。',
    'Moderate before channel selection': '選擇渠道前審核',
    'Scan prompts before selecting an upstream channel. This may add latency.':
      '在選擇上游渠道前掃描提示詞，可能增加延遲。',
    'Moderation base URL': '審核端點位址',
    'The endpoint should expose POST /moderations.':
      '端點應提供 POST /moderations 路徑。',
    'Moderation model': '審核模型',
    'Defaults to omni-moderation-latest when blank.':
      '留空時使用 omni-moderation-latest。',
    'Moderation API key': '審核 API 金鑰',
    'The key is write-only and is never shown after saving.':
      '金鑰僅可寫入，儲存後不會顯示。',
    'Moderation alert email': '審核告警信箱',
    'Optional alert recipient': '可選的告警收件信箱',
    'Receive an email after repeated moderation upstream failures.':
      '審核上游連續失敗後接收電子郵件通知。',
    'Moderation alert threshold': '審核告警閾值',
    'Failures within 30 minutes before an alert is sent.':
      '30 分鐘內達到此失敗次數後發送告警。',
    'Moderation cache TTL': '審核快取時間',
    'Seconds to reuse a successful moderation result.':
      '成功審核結果的重用時間（秒）。',
  },
  fr: {
    'Content Moderation': 'Modération du contenu',
    'Save moderation settings': 'Enregistrer les paramètres de modération',
    'Enable content moderation': 'Activer la modération du contenu',
    'Use an OpenAI-compatible moderation endpoint to scan user prompts.':
      'Utiliser un endpoint de modération compatible OpenAI pour analyser les prompts utilisateur.',
    'Moderate before channel selection': 'Modérer avant la sélection du canal',
    'Scan prompts before selecting an upstream channel. This may add latency.':
      'Analyser les prompts avant de sélectionner un canal amont. Cela peut ajouter de la latence.',
    'Moderation base URL': 'URL de base de modération',
    'The endpoint should expose POST /moderations.':
      'L’endpoint doit fournir POST /moderations.',
    'Moderation model': 'Modèle de modération',
    'Defaults to omni-moderation-latest when blank.':
      'Utilise omni-moderation-latest si le champ est vide.',
    'Moderation API key': 'Clé API de modération',
    'The key is write-only and is never shown after saving.':
      'La clé est en écriture seule et ne sera jamais affichée après l’enregistrement.',
    'Moderation alert email': 'E-mail d’alerte de modération',
    'Optional alert recipient': 'Destinataire d’alerte facultatif',
    'Receive an email after repeated moderation upstream failures.':
      'Recevoir un e-mail après des échecs répétés du service de modération amont.',
    'Moderation alert threshold': 'Seuil d’alerte de modération',
    'Failures within 30 minutes before an alert is sent.':
      'Nombre d’échecs en 30 minutes avant l’envoi d’une alerte.',
    'Moderation cache TTL': 'TTL du cache de modération',
    'Seconds to reuse a successful moderation result.':
      'Nombre de secondes pendant lesquelles un résultat réussi est réutilisé.',
  },
  ja: {
    'Content Moderation': 'コンテンツモデレーション',
    'Save moderation settings': 'モデレーション設定を保存',
    'Enable content moderation': 'コンテンツモデレーションを有効化',
    'Use an OpenAI-compatible moderation endpoint to scan user prompts.':
      'OpenAI 互換のモデレーションエンドポイントでユーザープロンプトを検査します。',
    'Moderate before channel selection': 'チャネル選択前にモデレーション',
    'Scan prompts before selecting an upstream channel. This may add latency.':
      '上流チャネルを選択する前にプロンプトを検査します。遅延が増える場合があります。',
    'Moderation base URL': 'モデレーションベース URL',
    'The endpoint should expose POST /moderations.':
      'エンドポイントは POST /moderations を提供する必要があります。',
    'Moderation model': 'モデレーションモデル',
    'Defaults to omni-moderation-latest when blank.':
      '空欄の場合は omni-moderation-latest を使用します。',
    'Moderation API key': 'モデレーション API キー',
    'The key is write-only and is never shown after saving.':
      'キーは書き込み専用で、保存後に表示されることはありません。',
    'Moderation alert email': 'モデレーションアラートメール',
    'Optional alert recipient': '任意のアラート受信者',
    'Receive an email after repeated moderation upstream failures.':
      'モデレーション上流で障害が繰り返されたときにメールを受信します。',
    'Moderation alert threshold': 'モデレーションアラートしきい値',
    'Failures within 30 minutes before an alert is sent.':
      'アラートを送信するまでの 30 分間の失敗回数です。',
    'Moderation cache TTL': 'モデレーションキャッシュ TTL',
    'Seconds to reuse a successful moderation result.':
      '成功したモデレーション結果を再利用する秒数です。',
  },
  ru: {
    'Content Moderation': 'Модерация контента',
    'Save moderation settings': 'Сохранить настройки модерации',
    'Enable content moderation': 'Включить модерацию контента',
    'Use an OpenAI-compatible moderation endpoint to scan user prompts.':
      'Использовать совместимый с OpenAI endpoint модерации для проверки пользовательских запросов.',
    'Moderate before channel selection': 'Модерировать до выбора канала',
    'Scan prompts before selecting an upstream channel. This may add latency.':
      'Проверять запросы до выбора upstream-канала. Это может увеличить задержку.',
    'Moderation base URL': 'Базовый URL модерации',
    'The endpoint should expose POST /moderations.':
      'Endpoint должен предоставлять POST /moderations.',
    'Moderation model': 'Модель модерации',
    'Defaults to omni-moderation-latest when blank.':
      'Если поле пусто, используется omni-moderation-latest.',
    'Moderation API key': 'API-ключ модерации',
    'The key is write-only and is never shown after saving.':
      'Ключ доступен только для записи и не отображается после сохранения.',
    'Moderation alert email': 'Почта для оповещений модерации',
    'Optional alert recipient': 'Необязательный получатель оповещений',
    'Receive an email after repeated moderation upstream failures.':
      'Получать письмо после повторяющихся сбоев upstream-модерации.',
    'Moderation alert threshold': 'Порог оповещения модерации',
    'Failures within 30 minutes before an alert is sent.':
      'Число сбоев за 30 минут до отправки оповещения.',
    'Moderation cache TTL': 'TTL кэша модерации',
    'Seconds to reuse a successful moderation result.':
      'Количество секунд для повторного использования успешного результата.',
  },
  vi: {
    'Content Moderation': 'Kiểm duyệt nội dung',
    'Save moderation settings': 'Lưu cài đặt kiểm duyệt',
    'Enable content moderation': 'Bật kiểm duyệt nội dung',
    'Use an OpenAI-compatible moderation endpoint to scan user prompts.':
      'Dùng endpoint kiểm duyệt tương thích OpenAI để quét prompt của người dùng.',
    'Moderate before channel selection': 'Kiểm duyệt trước khi chọn kênh',
    'Scan prompts before selecting an upstream channel. This may add latency.':
      'Quét prompt trước khi chọn kênh upstream. Điều này có thể làm tăng độ trễ.',
    'Moderation base URL': 'URL cơ sở kiểm duyệt',
    'The endpoint should expose POST /moderations.':
      'Endpoint phải cung cấp POST /moderations.',
    'Moderation model': 'Mô hình kiểm duyệt',
    'Defaults to omni-moderation-latest when blank.':
      'Mặc định dùng omni-moderation-latest khi để trống.',
    'Moderation API key': 'API key kiểm duyệt',
    'The key is write-only and is never shown after saving.':
      'Key chỉ được ghi và không bao giờ hiển thị sau khi lưu.',
    'Moderation alert email': 'Email cảnh báo kiểm duyệt',
    'Optional alert recipient': 'Người nhận cảnh báo (không bắt buộc)',
    'Receive an email after repeated moderation upstream failures.':
      'Nhận email khi upstream kiểm duyệt thất bại nhiều lần.',
    'Moderation alert threshold': 'Ngưỡng cảnh báo kiểm duyệt',
    'Failures within 30 minutes before an alert is sent.':
      'Số lần thất bại trong 30 phút trước khi gửi cảnh báo.',
    'Moderation cache TTL': 'TTL bộ nhớ đệm kiểm duyệt',
    'Seconds to reuse a successful moderation result.':
      'Số giây tái sử dụng kết quả kiểm duyệt thành công.',
  },
}

async function main() {
  for (const [locale, translations] of Object.entries(newKeys)) {
    const filePath = path.join(LOCALES_DIR, `${locale}.json`)
    const json = JSON.parse(await fs.readFile(filePath, 'utf8'))
    Object.assign(json.translation, translations)
    await fs.writeFile(filePath, stableStringify(json), 'utf8')
  }
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
})
