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
    'Official pricing': 'Official pricing',
    'Platform group pricing': 'Platform group pricing',
    'Remove amount {{amount}}': 'Remove amount {{amount}}',
    'Audit log cleanup': 'Audit log cleanup',
    'Audit log cleanup task started.': 'Audit log cleanup task started.',
    'Clean audit logs': 'Clean audit logs',
    'Permanently delete audit records older than the retention period. Usage logs and server log files are not affected.':
      'Permanently delete audit records older than the retention period. Usage logs and server log files are not affected.',
    'Days to retain': 'Days to retain',
    '{{count}} audit log entries removed.':
      '{{count}} audit log entries removed.',
    'No audit log entries matched the retention period.':
      'No audit log entries matched the retention period.',
    'Failed to clean audit logs': 'Failed to clean audit logs',
    'Enter a retention period between 1 and {{max}} days.':
      'Enter a retention period between 1 and {{max}} days.',
    'Audit log cleanup progress': 'Audit log cleanup progress',
    '{{processed}} of {{total}} audit log entries processed.':
      '{{processed}} of {{total}} audit log entries processed.',
    'Confirm audit log cleanup': 'Confirm audit log cleanup',
    'Audit records older than {{date}} will be permanently deleted. Only the audit_logs table is affected; usage logs and server log files will remain unchanged.':
      'Audit records older than {{date}} will be permanently deleted. Only the audit_logs table is affected; usage logs and server log files will remain unchanged.',
    'Delete audit logs': 'Delete audit logs',
    'Use responsibly; breaking limits is strictly prohibited!':
      'Use responsibly; breaking limits is strictly prohibited!',
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
    'Seconds to reuse sensitive-word and moderation results for identical user content.':
      'Seconds to reuse sensitive-word and moderation results for identical user content.',
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
    'Required moderation user IDs': 'Required moderation user IDs',
    'These users are always moderated, even when they or their group are exempt.':
      'These users are always moderated, even when they or their group are exempt.',
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
    'Test moderation connection': 'Test moderation connection',
    'View OpenAI usage statistics': 'View OpenAI usage statistics',
    'Enter a moderation base URL and API key first.':
      'Enter a moderation base URL and API key first.',
    'Moderation connection test failed.': 'Moderation connection test failed.',
    'Connection succeeded; the test text was flagged.':
      'Connection succeeded; the test text was flagged.',
    'Connection succeeded; the test text was allowed.':
      'Connection succeeded; the test text was allowed.',
    'Combined rules cannot contain empty keywords':
      'Combined rules cannot contain empty keywords',
    'Each combined rule can contain at most 5 keywords':
      'Each combined rule can contain at most 5 keywords',
    'Enter one keyword or combined rule per line':
      'Enter one keyword or combined rule per line',
    'Enter one rule per line. Use | to require all keywords in a combined rule (up to 5). Leave blank to disable the list.':
      'Enter one rule per line. Use | to require all keywords in a combined rule (up to 5). Leave blank to disable the list.',
    'Sensitive word hits': 'Sensitive word hits',
    'View matched sensitive words': 'View matched sensitive words',
    'Matched sensitive words': 'Matched sensitive words',
    'View the configured sensitive words matched by a request.':
      'View the configured sensitive words matched by a request.',
  },
  zh: {
    'Official pricing': '官方定价',
    'Platform group pricing': '平台分组定价',
    'Remove amount {{amount}}': '移除金额 {{amount}}',
    'Audit log cleanup': '审计日志清理',
    'Audit log cleanup task started.': '审计日志清理任务已启动。',
    'Clean audit logs': '清理审计日志',
    'Permanently delete audit records older than the retention period. Usage logs and server log files are not affected.':
      '永久删除早于保留期限的审计记录。用量日志和服务器日志文件不受影响。',
    'Days to retain': '保留天数',
    '{{count}} audit log entries removed.': '已删除 {{count}} 条审计日志。',
    'No audit log entries matched the retention period.':
      '没有符合保留期限的审计日志。',
    'Failed to clean audit logs': '清理审计日志失败',
    'Enter a retention period between 1 and {{max}} days.':
      '请输入 1 到 {{max}} 天之间的保留期限。',
    'Audit log cleanup progress': '审计日志清理进度',
    '{{processed}} of {{total}} audit log entries processed.':
      '已处理 {{processed}} / {{total}} 条审计日志。',
    'Confirm audit log cleanup': '确认清理审计日志',
    'Audit records older than {{date}} will be permanently deleted. Only the audit_logs table is affected; usage logs and server log files will remain unchanged.':
      '将永久删除 {{date}} 之前的审计记录。仅影响 audit_logs 表，用量日志和服务器日志文件将保持不变。',
    'Delete audit logs': '删除审计日志',
    'Use responsibly; breaking limits is strictly prohibited!':
      '文明使用，严禁破限！',
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
    'Seconds to reuse sensitive-word and moderation results for identical user content.':
      '相同用户内容复用敏感词和审核结果的秒数。',
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
    'Required moderation user IDs': '必须审核的用户 ID',
    'These users are always moderated, even when they or their group are exempt.':
      '这些用户始终需要审核，即使其自身或所属用户组已被豁免。',
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
    'Test moderation connection': '测试审核接口连接',
    'View OpenAI usage statistics': '查看 OpenAI 使用统计',
    'Enter a moderation base URL and API key first.':
      '请先填写审核接口地址和 API 密钥。',
    'Moderation connection test failed.': '审核接口连接测试失败。',
    'Connection succeeded; the test text was flagged.':
      '连接成功，测试文本被判定为违规。',
    'Connection succeeded; the test text was allowed.':
      '连接成功，测试文本通过审核。',
    'Combined rules cannot contain empty keywords': '组合规则不能包含空关键词',
    'Each combined rule can contain at most 5 keywords':
      '每条组合规则最多可包含 5 个关键词',
    'Enter one keyword or combined rule per line':
      '每行输入一个关键词或组合规则',
    'Enter one rule per line. Use | to require all keywords in a combined rule (up to 5). Leave blank to disable the list.':
      '每行输入一条规则。使用 | 分隔组合规则中必须同时出现的关键词（最多 5 个）。留空可禁用敏感词列表。',
    'Sensitive word hits': '敏感词命中次数',
    'View matched sensitive words': '查看命中的敏感词',
    'Matched sensitive words': '命中的敏感词',
    'View the configured sensitive words matched by a request.':
      '查看请求命中的已配置敏感词。',
  },
  'zh-TW': {
    'Official pricing': '官方定價',
    'Platform group pricing': '平台分組定價',
    'Remove amount {{amount}}': '移除金額 {{amount}}',
    'Audit log cleanup': '稽核日誌清理',
    'Audit log cleanup task started.': '稽核日誌清理工作已啟動。',
    'Clean audit logs': '清理稽核日誌',
    'Permanently delete audit records older than the retention period. Usage logs and server log files are not affected.':
      '永久刪除早於保留期限的稽核記錄。用量日誌與伺服器日誌檔案不受影響。',
    'Days to retain': '保留天數',
    '{{count}} audit log entries removed.': '已刪除 {{count}} 筆稽核日誌。',
    'No audit log entries matched the retention period.':
      '沒有符合保留期限的稽核日誌。',
    'Failed to clean audit logs': '清理稽核日誌失敗',
    'Enter a retention period between 1 and {{max}} days.':
      '請輸入 1 到 {{max}} 天之間的保留期限。',
    'Audit log cleanup progress': '稽核日誌清理進度',
    '{{processed}} of {{total}} audit log entries processed.':
      '已處理 {{processed}} / {{total}} 筆稽核日誌。',
    'Confirm audit log cleanup': '確認清理稽核日誌',
    'Audit records older than {{date}} will be permanently deleted. Only the audit_logs table is affected; usage logs and server log files will remain unchanged.':
      '將永久刪除 {{date}} 之前的稽核記錄。僅影響 audit_logs 資料表，用量日誌與伺服器日誌檔案將保持不變。',
    'Delete audit logs': '刪除稽核日誌',
    'Use responsibly; breaking limits is strictly prohibited!':
      '請文明使用，嚴禁突破限制！',
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
    'Seconds to reuse sensitive-word and moderation results for identical user content.':
      '相同使用者內容重用敏感詞與審核結果的秒數。',
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
    'Required moderation user IDs': '必須審核的使用者 ID',
    'These users are always moderated, even when they or their group are exempt.':
      '這些使用者一律接受審核，即使其本身或所屬使用者群組已獲豁免。',
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
    'Test moderation connection': '審核連線測試',
    'View OpenAI usage statistics': '查看 OpenAI 使用統計',
    'Enter a moderation base URL and API key first.':
      '請先填寫審核端點位址和 API 金鑰。',
    'Moderation connection test failed.': '審核連線測試失敗。',
    'Connection succeeded; the test text was flagged.':
      '連線成功，測試文字被判定為違規。',
    'Connection succeeded; the test text was allowed.':
      '連線成功，測試文字通過審核。',
    'Combined rules cannot contain empty keywords': '組合規則不能包含空關鍵詞',
    'Each combined rule can contain at most 5 keywords':
      '每條組合規則最多可包含 5 個關鍵詞',
    'Enter one keyword or combined rule per line':
      '每行輸入一個關鍵詞或組合規則',
    'Enter one rule per line. Use | to require all keywords in a combined rule (up to 5). Leave blank to disable the list.':
      '每行輸入一條規則。使用 | 分隔組合規則中必須同時出現的關鍵詞（最多 5 個）。留空可停用敏感詞清單。',
    'Sensitive word hits': '敏感詞命中次數',
    'View matched sensitive words': '查看命中的敏感詞',
    'Matched sensitive words': '命中的敏感詞',
    'View the configured sensitive words matched by a request.':
      '查看請求命中的已設定敏感詞。',
  },
  fr: {
    'Official pricing': 'Tarification officielle',
    'Platform group pricing': 'Tarification par groupe de la plateforme',
    'Remove amount {{amount}}': 'Supprimer le montant {{amount}}',
    'Audit log cleanup': "Nettoyage du journal d'audit",
    'Audit log cleanup task started.':
      "Tâche de nettoyage du journal d'audit démarrée.",
    'Clean audit logs': "Nettoyer le journal d'audit",
    'Permanently delete audit records older than the retention period. Usage logs and server log files are not affected.':
      "Supprime définitivement les enregistrements d'audit antérieurs à la durée de conservation. Les journaux d'utilisation et fichiers journaux du serveur ne sont pas affectés.",
    'Days to retain': 'Jours à conserver',
    '{{count}} audit log entries removed.':
      "{{count}} entrées du journal d'audit supprimées.",
    'No audit log entries matched the retention period.':
      "Aucune entrée du journal d'audit ne correspond à la durée de conservation.",
    'Failed to clean audit logs': "Échec du nettoyage du journal d'audit",
    'Enter a retention period between 1 and {{max}} days.':
      'Saisissez une durée de conservation comprise entre 1 et {{max}} jours.',
    'Audit log cleanup progress': "Progression du nettoyage du journal d'audit",
    '{{processed}} of {{total}} audit log entries processed.':
      "{{processed}} entrées du journal d'audit traitées sur {{total}}.",
    'Confirm audit log cleanup': "Confirmer le nettoyage du journal d'audit",
    'Audit records older than {{date}} will be permanently deleted. Only the audit_logs table is affected; usage logs and server log files will remain unchanged.':
      "Les enregistrements d'audit antérieurs au {{date}} seront définitivement supprimés. Seule la table audit_logs est concernée ; les journaux d'utilisation et fichiers journaux du serveur resteront inchangés.",
    'Delete audit logs': "Supprimer le journal d'audit",
    'Use responsibly; breaking limits is strictly prohibited!':
      'Utilisation responsable exigée. Contourner les limites est interdit !',
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
    'Seconds to reuse sensitive-word and moderation results for identical user content.':
      'Durée de réutilisation des résultats des mots sensibles et de modération pour un contenu identique.',
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
    'Required moderation user IDs': 'ID utilisateur à modération obligatoire',
    'These users are always moderated, even when they or their group are exempt.':
      'Ces utilisateurs sont toujours modérés, même s’ils sont exemptés ou si leur groupe l’est.',
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
    'Test moderation connection': 'Tester la connexion de modération',
    'View OpenAI usage statistics':
      'Voir les statistiques d’utilisation OpenAI',
    'Enter a moderation base URL and API key first.':
      'Saisissez d’abord une URL de base et une clé API de modération.',
    'Moderation connection test failed.':
      'Échec du test de connexion de modération.',
    'Connection succeeded; the test text was flagged.':
      'Connexion réussie ; le texte de test a été signalé.',
    'Connection succeeded; the test text was allowed.':
      'Connexion réussie ; le texte de test a été autorisé.',
    'Combined rules cannot contain empty keywords':
      'Les règles combinées ne peuvent pas contenir de mots-clés vides',
    'Each combined rule can contain at most 5 keywords':
      'Chaque règle combinée peut contenir au maximum 5 mots-clés',
    'Enter one keyword or combined rule per line':
      'Saisissez un mot-clé ou une règle combinée par ligne',
    'Enter one rule per line. Use | to require all keywords in a combined rule (up to 5). Leave blank to disable the list.':
      'Saisissez une règle par ligne. Utilisez | pour exiger tous les mots-clés d’une règle combinée (5 maximum). Laissez vide pour désactiver la liste.',
    'Sensitive word hits': 'Détections de mots sensibles',
    'View matched sensitive words': 'Voir les mots sensibles détectés',
    'Matched sensitive words': 'Mots sensibles détectés',
    'View the configured sensitive words matched by a request.':
      'Voir les mots sensibles configurés détectés dans une requête.',
  },
  ja: {
    'Official pricing': '公式価格',
    'Platform group pricing': 'プラットフォームのグループ価格',
    'Remove amount {{amount}}': '金額 {{amount}} を削除',
    'Audit log cleanup': '監査ログのクリーンアップ',
    'Audit log cleanup task started.':
      '監査ログのクリーンアップを開始しました。',
    'Clean audit logs': '監査ログをクリーンアップ',
    'Permanently delete audit records older than the retention period. Usage logs and server log files are not affected.':
      '保持期間より古い監査記録を完全に削除します。使用量ログとサーバーログファイルには影響しません。',
    'Days to retain': '保持日数',
    '{{count}} audit log entries removed.':
      '{{count}} 件の監査ログを削除しました。',
    'No audit log entries matched the retention period.':
      '保持期間より古い監査ログはありませんでした。',
    'Failed to clean audit logs': '監査ログのクリーンアップに失敗しました',
    'Enter a retention period between 1 and {{max}} days.':
      '保持期間を 1～{{max}} 日で入力してください。',
    'Audit log cleanup progress': '監査ログのクリーンアップ進捗',
    '{{processed}} of {{total}} audit log entries processed.':
      '{{total}} 件中 {{processed}} 件の監査ログを処理しました。',
    'Confirm audit log cleanup': '監査ログのクリーンアップを確認',
    'Audit records older than {{date}} will be permanently deleted. Only the audit_logs table is affected; usage logs and server log files will remain unchanged.':
      '{{date}} より古い監査記録は完全に削除されます。影響するのは audit_logs テーブルのみで、使用量ログとサーバーログファイルは変更されません。',
    'Delete audit logs': '監査ログを削除',
    'Use responsibly; breaking limits is strictly prohibited!':
      '責任を持って利用し、制限の突破は固く禁止します！',
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
    'Seconds to reuse sensitive-word and moderation results for identical user content.':
      '同一のユーザー内容に対するセンシティブワードとモデレーションの結果を再利用する秒数です。',
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
    'Required moderation user IDs': '常に審査するユーザー ID',
    'These users are always moderated, even when they or their group are exempt.':
      'これらのユーザーは、本人または所属グループが除外対象でも常に審査されます。',
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
    'Test moderation connection': 'モデレーション接続をテスト',
    'View OpenAI usage statistics': 'OpenAI の利用統計を表示',
    'Enter a moderation base URL and API key first.':
      '先にモデレーションのベース URL と API キーを入力してください。',
    'Moderation connection test failed.':
      'モデレーション接続テストに失敗しました。',
    'Connection succeeded; the test text was flagged.':
      '接続に成功しました。テスト文は違反として判定されました。',
    'Connection succeeded; the test text was allowed.':
      '接続に成功しました。テスト文は許可されました。',
    'Combined rules cannot contain empty keywords':
      '組み合わせルールに空のキーワードを含めることはできません',
    'Each combined rule can contain at most 5 keywords':
      '各組み合わせルールに設定できるキーワードは最大5個です',
    'Enter one keyword or combined rule per line':
      '1行に1つのキーワードまたは組み合わせルールを入力',
    'Enter one rule per line. Use | to require all keywords in a combined rule (up to 5). Leave blank to disable the list.':
      '1行に1つのルールを入力します。組み合わせルールですべてのキーワードを必須にするには、| で区切ります（最大5個）。空欄にするとリストが無効になります。',
    'Sensitive word hits': '機密語の検出回数',
    'View matched sensitive words': '検出された機密語を表示',
    'Matched sensitive words': '検出された機密語',
    'View the configured sensitive words matched by a request.':
      'リクエストで検出された設定済みの機密語を表示します。',
  },
  ru: {
    'Official pricing': 'Официальная цена',
    'Platform group pricing': 'Групповая цена платформы',
    'Remove amount {{amount}}': 'Удалить сумму {{amount}}',
    'Audit log cleanup': 'Очистка журнала аудита',
    'Audit log cleanup task started.':
      'Задача очистки журнала аудита запущена.',
    'Clean audit logs': 'Очистить журнал аудита',
    'Permanently delete audit records older than the retention period. Usage logs and server log files are not affected.':
      'Безвозвратно удаляет записи аудита старше срока хранения. Журналы использования и файлы журналов сервера не затрагиваются.',
    'Days to retain': 'Дней хранения',
    '{{count}} audit log entries removed.':
      'Удалено записей журнала аудита: {{count}}.',
    'No audit log entries matched the retention period.':
      'Записей журнала аудита старше срока хранения не найдено.',
    'Failed to clean audit logs': 'Не удалось очистить журнал аудита',
    'Enter a retention period between 1 and {{max}} days.':
      'Укажите срок хранения от 1 до {{max}} дней.',
    'Audit log cleanup progress': 'Ход очистки журнала аудита',
    '{{processed}} of {{total}} audit log entries processed.':
      'Обработано записей журнала аудита: {{processed}} из {{total}}.',
    'Confirm audit log cleanup': 'Подтвердите очистку журнала аудита',
    'Audit records older than {{date}} will be permanently deleted. Only the audit_logs table is affected; usage logs and server log files will remain unchanged.':
      'Записи аудита старше {{date}} будут безвозвратно удалены. Изменяется только таблица audit_logs; журналы использования и файлы журналов сервера останутся без изменений.',
    'Delete audit logs': 'Удалить журнал аудита',
    'Use responsibly; breaking limits is strictly prohibited!':
      'Используйте сервис ответственно; обход ограничений строго запрещён!',
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
    'Seconds to reuse sensitive-word and moderation results for identical user content.':
      'Срок повторного использования результатов проверки чувствительных слов и модерации для одинакового содержимого.',
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
    'Required moderation user IDs':
      'ID пользователей для обязательной модерации',
    'These users are always moderated, even when they or their group are exempt.':
      'Эти пользователи всегда проходят модерацию, даже если они или их группа освобождены от неё.',
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
    'Test moderation connection': 'Проверить подключение модерации',
    'View OpenAI usage statistics':
      'Просмотреть статистику использования OpenAI',
    'Enter a moderation base URL and API key first.':
      'Сначала укажите базовый URL и API-ключ модерации.',
    'Moderation connection test failed.':
      'Не удалось проверить подключение модерации.',
    'Connection succeeded; the test text was flagged.':
      'Подключение успешно; тестовый текст отмечен как нарушающий правила.',
    'Connection succeeded; the test text was allowed.':
      'Подключение успешно; тестовый текст разрешён.',
    'Combined rules cannot contain empty keywords':
      'Комбинированные правила не могут содержать пустые ключевые слова',
    'Each combined rule can contain at most 5 keywords':
      'Каждое комбинированное правило может содержать не более 5 ключевых слов',
    'Enter one keyword or combined rule per line':
      'Вводите по одному ключевому слову или комбинированному правилу в строке',
    'Enter one rule per line. Use | to require all keywords in a combined rule (up to 5). Leave blank to disable the list.':
      'Вводите по одному правилу в строке. Используйте |, чтобы потребовать все ключевые слова в комбинированном правиле (до 5). Оставьте поле пустым, чтобы отключить список.',
    'Sensitive word hits': 'Срабатывания по чувствительным словам',
    'View matched sensitive words': 'Просмотр совпавших чувствительных слов',
    'Matched sensitive words': 'Совпавшие чувствительные слова',
    'View the configured sensitive words matched by a request.':
      'Просматривать настроенные чувствительные слова, совпавшие в запросе.',
  },
  vi: {
    'Official pricing': 'Giá chính thức',
    'Platform group pricing': 'Giá theo nhóm nền tảng',
    'Remove amount {{amount}}': 'Xóa số tiền {{amount}}',
    'Audit log cleanup': 'Dọn dẹp nhật ký kiểm tra',
    'Audit log cleanup task started.':
      'Đã bắt đầu tác vụ dọn dẹp nhật ký kiểm tra.',
    'Clean audit logs': 'Dọn dẹp nhật ký kiểm tra',
    'Permanently delete audit records older than the retention period. Usage logs and server log files are not affected.':
      'Xóa vĩnh viễn các bản ghi kiểm tra cũ hơn thời hạn lưu giữ. Nhật ký sử dụng và tệp nhật ký máy chủ không bị ảnh hưởng.',
    'Days to retain': 'Số ngày lưu giữ',
    '{{count}} audit log entries removed.':
      'Đã xóa {{count}} mục nhật ký kiểm tra.',
    'No audit log entries matched the retention period.':
      'Không có mục nhật ký kiểm tra nào cũ hơn thời hạn lưu giữ.',
    'Failed to clean audit logs': 'Không thể dọn dẹp nhật ký kiểm tra',
    'Enter a retention period between 1 and {{max}} days.':
      'Nhập thời hạn lưu giữ từ 1 đến {{max}} ngày.',
    'Audit log cleanup progress': 'Tiến độ dọn dẹp nhật ký kiểm tra',
    '{{processed}} of {{total}} audit log entries processed.':
      'Đã xử lý {{processed}} trên {{total}} mục nhật ký kiểm tra.',
    'Confirm audit log cleanup': 'Xác nhận dọn dẹp nhật ký kiểm tra',
    'Audit records older than {{date}} will be permanently deleted. Only the audit_logs table is affected; usage logs and server log files will remain unchanged.':
      'Các bản ghi kiểm tra cũ hơn {{date}} sẽ bị xóa vĩnh viễn. Chỉ bảng audit_logs bị ảnh hưởng; nhật ký sử dụng và tệp nhật ký máy chủ vẫn được giữ nguyên.',
    'Delete audit logs': 'Xóa nhật ký kiểm tra',
    'Use responsibly; breaking limits is strictly prohibited!':
      'Vui lòng sử dụng có trách nhiệm; nghiêm cấm vượt qua giới hạn!',
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
    'Seconds to reuse sensitive-word and moderation results for identical user content.':
      'Số giây tái sử dụng kết quả từ nhạy cảm và kiểm duyệt cho cùng nội dung người dùng.',
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
    'Required moderation user IDs': 'ID người dùng luôn phải kiểm duyệt',
    'These users are always moderated, even when they or their group are exempt.':
      'Những người dùng này luôn được kiểm duyệt, ngay cả khi họ hoặc nhóm của họ được miễn.',
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
    'Test moderation connection': 'Kiểm tra kết nối kiểm duyệt',
    'View OpenAI usage statistics': 'Xem thống kê sử dụng OpenAI',
    'Enter a moderation base URL and API key first.':
      'Vui lòng nhập URL cơ sở và API key kiểm duyệt trước.',
    'Moderation connection test failed.':
      'Kiểm tra kết nối kiểm duyệt thất bại.',
    'Connection succeeded; the test text was flagged.':
      'Kết nối thành công; văn bản kiểm tra bị đánh dấu vi phạm.',
    'Connection succeeded; the test text was allowed.':
      'Kết nối thành công; văn bản kiểm tra được cho phép.',
    'Combined rules cannot contain empty keywords':
      'Quy tắc kết hợp không được chứa từ khóa trống',
    'Each combined rule can contain at most 5 keywords':
      'Mỗi quy tắc kết hợp chỉ được chứa tối đa 5 từ khóa',
    'Enter one keyword or combined rule per line':
      'Nhập một từ khóa hoặc quy tắc kết hợp trên mỗi dòng',
    'Enter one rule per line. Use | to require all keywords in a combined rule (up to 5). Leave blank to disable the list.':
      'Nhập một quy tắc trên mỗi dòng. Dùng | để yêu cầu tất cả từ khóa trong quy tắc kết hợp (tối đa 5). Để trống để tắt danh sách.',
    'Sensitive word hits': 'Số lần phát hiện từ ngữ nhạy cảm',
    'View matched sensitive words': 'Xem các từ nhạy cảm đã khớp',
    'Matched sensitive words': 'Các từ nhạy cảm đã khớp',
    'View the configured sensitive words matched by a request.':
      'Xem các từ nhạy cảm đã cấu hình khớp với yêu cầu.',
  },
}

const removedKeys = [
  'Remove ${{amount}}',
  'Seconds to reuse a successful moderation result.',
  'Sensitive word checks',
  'Sensitive word hit rate',
]

async function main() {
  for (const [locale, translations] of Object.entries(newKeys)) {
    const filePath = path.join(LOCALES_DIR, `${locale}.json`)
    const json = JSON.parse(await fs.readFile(filePath, 'utf8'))
    Object.assign(json.translation, translations)
    for (const key of removedKeys) {
      delete json.translation[key]
    }
    json.translation = Object.fromEntries(
      Object.entries(json.translation).sort(([a], [b]) => a.localeCompare(b))
    )
    await fs.writeFile(filePath, stableStringify(json), 'utf8')
  }
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
})
