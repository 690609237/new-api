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
    'Moderation sample rate': 'Moderation sample rate',
    'Percentage of non-exempt users selected for moderation. 100% checks everyone.':
      'Percentage of non-exempt users selected for moderation. 100% checks everyone.',
    'Exempt user IDs': 'Exempt user IDs',
    'One user ID per line': 'One user ID per line',
    'These users bypass moderation. Commas and line breaks are supported.':
      'These users bypass moderation. Commas and line breaks are supported.',
    'Exempt user groups': 'Exempt user groups',
    'One group per line': 'One group per line',
    'Users in these groups bypass moderation. Matching is case-insensitive.':
      'Users in these groups bypass moderation. Matching is case-insensitive.',
    'Moderation Audit': 'Moderation Audit',
    Decision: 'Decision',
    'Result source': 'Result source',
    'Cache hit': 'Cache hit',
    'API request': 'API request',
    Flagged: 'Flagged',
    Allowed: 'Allowed',
    'Submitted content': 'Submitted content',
    'Supported email domains: gmail.com, 163.com, 126.com, qq.com, outlook.com, hotmail.com, icloud.com, yahoo.com, foxmail.com, yeah.net, aliyun.com, sina.com, sina.cn, sohu.com, tom.com, 21cn.com':
      'Supported email domains: gmail.com, 163.com, 126.com, qq.com, outlook.com, hotmail.com, icloud.com, yahoo.com, foxmail.com, yeah.net, aliyun.com, sina.com, sina.cn, sohu.com, tom.com, 21cn.com',
    "If you don't receive the verification code, check your spam folder. If you still can't find it, please contact the author.":
      "If you don't receive the verification code, check your spam folder. If you still can't find it, please contact the author.",
    'Violation limit updated successfully':
      'Violation limit updated successfully',
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
    'Moderation sample rate': '审核采样比例',
    'Percentage of non-exempt users selected for moderation. 100% checks everyone.':
      '未豁免用户中参与审核的比例，100% 表示全部审核。',
    'Exempt user IDs': '豁免用户 ID',
    'One user ID per line': '每行一个用户 ID',
    'These users bypass moderation. Commas and line breaks are supported.':
      '这些用户会跳过审核，支持使用逗号或换行分隔。',
    'Exempt user groups': '豁免用户组',
    'One group per line': '每行一个用户组',
    'Users in these groups bypass moderation. Matching is case-insensitive.':
      '这些用户组中的用户会跳过审核，匹配不区分大小写。',
    'Moderation Audit': '审核追溯',
    Decision: '审核结果',
    'Result source': '结果来源',
    'Cache hit': '命中缓存',
    'API request': '调用审核接口',
    Flagged: '违规',
    Allowed: '通过',
    'Submitted content': '送审内容',
    'Supported email domains: gmail.com, 163.com, 126.com, qq.com, outlook.com, hotmail.com, icloud.com, yahoo.com, foxmail.com, yeah.net, aliyun.com, sina.com, sina.cn, sohu.com, tom.com, 21cn.com':
      '支持的邮箱类型有：gmail.com、163.com、126.com、qq.com、outlook.com、hotmail.com、icloud.com、yahoo.com、foxmail.com、yeah.net、aliyun.com、sina.com、sina.cn、sohu.com、tom.com、21cn.com',
    "If you don't receive the verification code, check your spam folder. If you still can't find it, please contact the author.":
      '如果收不到验证码，请检查邮件垃圾箱；如果仍然找不到，请联系作者。',
    'Violation limit updated successfully': '违规限制更新成功',
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
    'Moderation sample rate': '審核採樣比例',
    'Percentage of non-exempt users selected for moderation. 100% checks everyone.':
      '未豁免使用者中參與審核的比例，100% 表示全部審核。',
    'Exempt user IDs': '豁免使用者 ID',
    'One user ID per line': '每行一個使用者 ID',
    'These users bypass moderation. Commas and line breaks are supported.':
      '這些使用者會跳過審核，支援使用逗號或換行分隔。',
    'Exempt user groups': '豁免使用者群組',
    'One group per line': '每行一個使用者群組',
    'Users in these groups bypass moderation. Matching is case-insensitive.':
      '這些使用者群組中的使用者會跳過審核，匹配不分大小寫。',
    'Moderation Audit': '審核追溯',
    Decision: '審核結果',
    'Result source': '結果來源',
    'Cache hit': '命中快取',
    'API request': '呼叫審核 API',
    Flagged: '違規',
    Allowed: '通過',
    'Submitted content': '送審內容',
    'Supported email domains: gmail.com, 163.com, 126.com, qq.com, outlook.com, hotmail.com, icloud.com, yahoo.com, foxmail.com, yeah.net, aliyun.com, sina.com, sina.cn, sohu.com, tom.com, 21cn.com':
      '支援的電子郵件類型有：gmail.com、163.com、126.com、qq.com、outlook.com、hotmail.com、icloud.com、yahoo.com、foxmail.com、yeah.net、aliyun.com、sina.com、sina.cn、sohu.com、tom.com、21cn.com',
    "If you don't receive the verification code, check your spam folder. If you still can't find it, please contact the author.":
      '如果收不到驗證碼，請檢查郵件垃圾箱；如果仍然找不到，請聯絡作者。',
    'Violation limit updated successfully': '違規限制更新成功',
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
    'Moderation sample rate': 'Taux d’échantillonnage de la modération',
    'Percentage of non-exempt users selected for moderation. 100% checks everyone.':
      'Pourcentage des utilisateurs non exemptés soumis à la modération. 100 % vérifie tout le monde.',
    'Exempt user IDs': 'IDs utilisateur exemptés',
    'One user ID per line': 'Un ID utilisateur par ligne',
    'These users bypass moderation. Commas and line breaks are supported.':
      'Ces utilisateurs contournent la modération. Les virgules et retours à la ligne sont acceptés.',
    'Exempt user groups': 'Groupes utilisateur exemptés',
    'One group per line': 'Un groupe par ligne',
    'Users in these groups bypass moderation. Matching is case-insensitive.':
      'Les utilisateurs de ces groupes contournent la modération. La correspondance ignore la casse.',
    'Moderation Audit': 'Audit de modération',
    Decision: 'Décision',
    'Result source': 'Source du résultat',
    'Cache hit': 'Cache utilisé',
    'API request': 'Requête API',
    Flagged: 'Signalé',
    Allowed: 'Autorisé',
    'Submitted content': 'Contenu soumis',
    'Supported email domains: gmail.com, 163.com, 126.com, qq.com, outlook.com, hotmail.com, icloud.com, yahoo.com, foxmail.com, yeah.net, aliyun.com, sina.com, sina.cn, sohu.com, tom.com, 21cn.com':
      'Domaines e-mail pris en charge : gmail.com, 163.com, 126.com, qq.com, outlook.com, hotmail.com, icloud.com, yahoo.com, foxmail.com, yeah.net, aliyun.com, sina.com, sina.cn, sohu.com, tom.com, 21cn.com',
    "If you don't receive the verification code, check your spam folder. If you still can't find it, please contact the author.":
      "Si vous ne recevez pas le code de vérification, vérifiez vos spams. Si vous ne le trouvez toujours pas, contactez l'auteur.",
    'Violation limit updated successfully': 'Limite de violations mise à jour',
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
    'Moderation sample rate': 'モデレーションサンプル率',
    'Percentage of non-exempt users selected for moderation. 100% checks everyone.':
      '除外されていないユーザーのうちモデレーション対象にする割合です。100% は全員を検査します。',
    'Exempt user IDs': '除外するユーザー ID',
    'One user ID per line': '1 行に 1 つのユーザー ID',
    'These users bypass moderation. Commas and line breaks are supported.':
      'これらのユーザーはモデレーションを回避します。カンマと改行に対応しています。',
    'Exempt user groups': '除外するユーザーグループ',
    'One group per line': '1 行に 1 つのグループ',
    'Users in these groups bypass moderation. Matching is case-insensitive.':
      'これらのグループのユーザーはモデレーションを回避します。大文字と小文字は区別しません。',
    'Moderation Audit': 'モデレーション監査',
    Decision: '判定',
    'Result source': '結果ソース',
    'Cache hit': 'キャッシュヒット',
    'API request': 'API リクエスト',
    Flagged: '違反',
    Allowed: '許可',
    'Submitted content': '送信内容',
    'Supported email domains: gmail.com, 163.com, 126.com, qq.com, outlook.com, hotmail.com, icloud.com, yahoo.com, foxmail.com, yeah.net, aliyun.com, sina.com, sina.cn, sohu.com, tom.com, 21cn.com':
      '対応しているメールドメイン：gmail.com、163.com、126.com、qq.com、outlook.com、hotmail.com、icloud.com、yahoo.com、foxmail.com、yeah.net、aliyun.com、sina.com、sina.cn、sohu.com、tom.com、21cn.com',
    "If you don't receive the verification code, check your spam folder. If you still can't find it, please contact the author.":
      '認証コードが届かない場合は迷惑メールフォルダーを確認してください。それでも見つからない場合は作者にお問い合わせください。',
    'Violation limit updated successfully': '違反上限を更新しました',
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
    'Moderation sample rate': 'Доля выборочной модерации',
    'Percentage of non-exempt users selected for moderation. 100% checks everyone.':
      'Процент пользователей без исключений, выбранных для модерации. 100% проверяет всех.',
    'Exempt user IDs': 'Идентификаторы исключённых пользователей',
    'One user ID per line': 'Один ID пользователя в строке',
    'These users bypass moderation. Commas and line breaks are supported.':
      'Эти пользователи пропускают модерацию. Поддерживаются запятые и переносы строк.',
    'Exempt user groups': 'Исключённые группы пользователей',
    'One group per line': 'Одна группа в строке',
    'Users in these groups bypass moderation. Matching is case-insensitive.':
      'Пользователи этих групп пропускают модерацию. Регистр не учитывается.',
    'Moderation Audit': 'Аудит модерации',
    Decision: 'Решение',
    'Result source': 'Источник результата',
    'Cache hit': 'Попадание в кэш',
    'API request': 'Запрос к API',
    Flagged: 'Нарушение',
    Allowed: 'Разрешено',
    'Submitted content': 'Отправленное содержимое',
    'Supported email domains: gmail.com, 163.com, 126.com, qq.com, outlook.com, hotmail.com, icloud.com, yahoo.com, foxmail.com, yeah.net, aliyun.com, sina.com, sina.cn, sohu.com, tom.com, 21cn.com':
      'Поддерживаемые почтовые домены: gmail.com, 163.com, 126.com, qq.com, outlook.com, hotmail.com, icloud.com, yahoo.com, foxmail.com, yeah.net, aliyun.com, sina.com, sina.cn, sohu.com, tom.com, 21cn.com',
    "If you don't receive the verification code, check your spam folder. If you still can't find it, please contact the author.":
      'Если вы не получили код подтверждения, проверьте папку «Спам». Если найти его не удалось, обратитесь к автору.',
    'Violation limit updated successfully': 'Лимит нарушений обновлён',
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
    'Moderation sample rate': 'Tỷ lệ lấy mẫu kiểm duyệt',
    'Percentage of non-exempt users selected for moderation. 100% checks everyone.':
      'Tỷ lệ người dùng không được miễn kiểm duyệt sẽ được chọn. 100% là kiểm tra tất cả.',
    'Exempt user IDs': 'ID người dùng được miễn',
    'One user ID per line': 'Mỗi dòng một ID người dùng',
    'These users bypass moderation. Commas and line breaks are supported.':
      'Những người dùng này sẽ bỏ qua kiểm duyệt. Hỗ trợ dấu phẩy và xuống dòng.',
    'Exempt user groups': 'Nhóm người dùng được miễn',
    'One group per line': 'Mỗi dòng một nhóm',
    'Users in these groups bypass moderation. Matching is case-insensitive.':
      'Người dùng trong các nhóm này sẽ bỏ qua kiểm duyệt. Không phân biệt hoa thường.',
    'Moderation Audit': 'Kiểm tra kiểm duyệt',
    Decision: 'Kết quả',
    'Result source': 'Nguồn kết quả',
    'Cache hit': 'Trúng bộ nhớ đệm',
    'API request': 'Yêu cầu API',
    Flagged: 'Vi phạm',
    Allowed: 'Được phép',
    'Submitted content': 'Nội dung đã gửi',
    'Supported email domains: gmail.com, 163.com, 126.com, qq.com, outlook.com, hotmail.com, icloud.com, yahoo.com, foxmail.com, yeah.net, aliyun.com, sina.com, sina.cn, sohu.com, tom.com, 21cn.com':
      'Các miền email được hỗ trợ: gmail.com, 163.com, 126.com, qq.com, outlook.com, hotmail.com, icloud.com, yahoo.com, foxmail.com, yeah.net, aliyun.com, sina.com, sina.cn, sohu.com, tom.com, 21cn.com',
    "If you don't receive the verification code, check your spam folder. If you still can't find it, please contact the author.":
      'Nếu không nhận được mã xác minh, hãy kiểm tra thư mục spam. Nếu vẫn không tìm thấy, vui lòng liên hệ tác giả.',
    'Violation limit updated successfully': 'Đã cập nhật giới hạn vi phạm',
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
