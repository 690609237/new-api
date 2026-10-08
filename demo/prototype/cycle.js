const DAY_MS = 86400000;
function dateDay(value) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value || '')) return NaN;
  const timestamp = Date.parse(`${value}T00:00:00Z`);
  return Number.isFinite(timestamp) && new Date(timestamp).toISOString().slice(0, 10) === value ? timestamp / DAY_MS : NaN;
}
function dayDate(day) { return new Date(day * DAY_MS).toISOString().slice(0, 10); }
function cycleFor(patient) {
  const plan = patient.plan;
  const starts = (patient.versions || [plan]).map(version => version.start).filter(value => Number.isFinite(dateDay(value))).sort();
  return {
    start: plan.cycle?.start || starts[0] || plan.start,
    end: plan.end,
    kind: plan.cycle?.kind || (patient.path === '检查后随访' ? '观察周期' : '康复计划'),
    goal: plan.cycle?.goal || '周期目标尚待医护补充，请先按已发布的阶段计划执行。',
    reviewEvery: plan.cycle?.reviewEvery || 0
  };
}
function cycleMetrics(cycle, today = TODAY) {
  const start = dateDay(cycle.start), end = dateDay(cycle.end), current = dateDay(today);
  const total = end - start + 1;
  const elapsed = Math.max(0, Math.min(total, current - start + 1));
  const interval = cycle.reviewEvery || total;
  const count = Math.ceil(total / interval);
  const currentIndex = Math.max(0, Math.min(count - 1, Math.floor((current - start) / interval)));
  return {start, end, total, elapsed, interval, count, currentIndex, before: current < start, after: current > end};
}
function cycleSegment(cycle, index) {
  const metrics = cycleMetrics(cycle);
  const start = metrics.start + index * metrics.interval;
  const end = Math.min(metrics.end, start + metrics.interval - 1);
  return {start: dayDate(start), end: dayDate(end)};
}
function cycleEditor(patient) {
  const cycle = cycleFor(patient);
  return `<section class="cycle-editor"><h3>患者可见的长期计划 / 观察周期</h3><p class="muted">周期起点与版本生效日分开保存，调整任务不会重新计算整个周期。</p>
    <div class="form-row">${field('周期类型', `<select id="cycle-kind"><option ${cycle.kind === '康复计划' ? 'selected' : ''}>康复计划</option><option ${cycle.kind === '观察周期' ? 'selected' : ''}>观察周期</option></select>`)}${field('周期开始日期', input('cycle-start', cycle.start, 'date', 'min="2025-01-01" max="2027-12-31"'))}</div>
    <div class="form-row">${field('计划 / 观察结束日期', input('plan-end', cycle.end, 'date', 'min="2025-01-01" max="2027-12-31"'))}${field('每隔多少天回访一次', input('cycle-interval', cycle.reviewEvery || 7, 'number', 'min="1" max="90"'), '按周期第 N 天安排首次回访；末段不足 N 天，在周期末回访。')}</div>
    ${field('本周期目标（患者可见）', textarea('cycle-goal', patient.plan.cycle?.goal || '', '例如：持续记录居家执行情况，在回访节点核对问题并讨论后续安排。'))}
    <div class="notice">日期和回访频率仅用于安排服务，不会自动升级训练、调整处方或判定康复。周期结束仍需人工决定后续安排。</div></section>`;
}
function readCycle() {
  return {kind: read('cycle-kind'), start: read('cycle-start'), reviewEvery: Number(read('cycle-interval')), goal: read('cycle-goal')};
}
function cycleValidation(cycle, end) {
  if (!['康复计划', '观察周期'].includes(cycle.kind)) return '请选择计划或观察周期类型。';
  if (!Number.isFinite(dateDay(cycle.start)) || !Number.isFinite(dateDay(end)) || cycle.start < '2025-01-01' || end > '2027-12-31' || cycle.start > end) return '请填写有效的周期起止日期，结束日期不得早于开始日期。';
  if (dateDay(end) - dateDay(cycle.start) + 1 > 366) return '本原型单个周期最长支持 366 天，请拆分后续观察周期。';
  if (!Number.isInteger(cycle.reviewEvery) || cycle.reviewEvery < 1 || cycle.reviewEvery > 90) return '回访间隔须为 1–90 的整数天。';
  if (cycle.goal.length < 6 || cycle.goal.length > 500) return '请填写 6–500 个字的周期目标。';
  return '';
}
function cycleIsActive(patient) {
  const cycle = cycleFor(patient);
  return TODAY >= cycle.start && TODAY <= cycle.end;
}
function cycleHomeLink(patient) {
  const cycle = cycleFor(patient), metrics = cycleMetrics(cycle);
  const status = metrics.before ? '尚未开始' : metrics.after ? '周期已到期 · 待医护评估' : `周期第 ${metrics.elapsed} / ${metrics.total} 天`;
  return `<button class="cycle-home-link" data-action="patient-page" data-value="cycle"><span>${icon('list')}</span><span><strong>查看整个${escapeHtml(cycle.kind)}</strong><small>${displayDate(cycle.start)} — ${displayDate(cycle.end)}<br>${status}</small></span>${icon('arrow')}</button>`;
}
function patientCycle(patient) {
  const cycle = cycleFor(patient), metrics = cycleMetrics(cycle);
  const selectionKey = `${patient.id}:${patient.plan.version}`;
  const selected = ui.cycleSelection?.key === selectionKey ? Math.max(0, Math.min(metrics.count - 1, ui.cycleSelection.index)) : metrics.currentIndex;
  const segment = cycleSegment(cycle, selected);
  const upcoming = segment.start > TODAY, past = segment.end < TODAY;
  const records = patientRecords(patient.id).filter(record => record.date >= segment.start && record.date <= segment.end);
  const daysWithRecords = new Set(records.map(record => record.date)).size;
  const nextReview = cycle.reviewEvery && !metrics.after ? cycleSegment(cycle, metrics.currentIndex).end : '';
  const firstVisible = Math.max(0, Math.min(metrics.count - 5, selected - 2));
  const status = metrics.before ? '等待周期开始' : metrics.after ? '周期已到期 · 待医护评估' : `正在第 ${metrics.elapsed} 天`;
  return `<div class="phone-page-title"><span class="eyebrow">MY CARE PLAN</span><h2>不只今天，看见整个周期</h2><p>由${escapeHtml(patient.plan.author)}设置 · 您的任务与回访安排</p></div>
    <section class="cycle-hero"><div class="between"><h3>我的${escapeHtml(cycle.kind)}</h3>${sourceTag(`V${patient.plan.version}`, 'green')}</div><p class="cycle-dates">${displayDate(cycle.start)} — ${displayDate(cycle.end)}</p><strong class="cycle-goal">${escapeHtml(cycle.goal)}</strong>
      <div class="cycle-stats"><div><b>${metrics.total}<small> 天</small></b><span>完整周期</span></div><div><b>${cycle.reviewEvery || '—'}<small>${cycle.reviewEvery ? ' 天' : ''}</small></b><span>${cycle.reviewEvery ? '回访一次' : '回访频率待设置'}</span></div></div>
      <div class="progress-track"><span style="width:${metrics.elapsed / metrics.total * 100}%"></span></div><div class="between"><small>${status}</small><small>时间进度，非恢复评分</small></div></section>
    <section class="cycle-next"><span>${icon('chat')}</span><div><h3>${nextReview ? `下个计划回访 · ${displayDate(nextReview)}` : metrics.after ? '等待医护确认下一步' : '回访节点待团队设置'}</h3><p>${nextReview ? '届时核对记录与不适反馈。此处仅展示计划，不代表已发出提醒或完成回访。' : metrics.after ? '周期到期不会自动结案、延长计划或关闭异常事项。' : '目前只有医护计划的起止日期，具体回访节奏待团队补充。'}</p></div></section>
    <div class="phone-section-title"><h3>分段查看安排</h3><small>${cycle.reviewEvery ? `每 ${cycle.reviewEvery} 天一段` : '当前完整周期'}</small></div>
    <p class="cycle-hint">这是回访分段，不是自动升级的治疗阶段。点击查看前后周期。</p>
    <div class="cycle-periods" role="group" aria-label="选择回访分段">${Array.from({length: Math.min(5, metrics.count)}, (_, offset) => {const index = firstVisible + offset;return `<button class="${index === selected ? 'active' : ''}" aria-pressed="${index === selected}" data-action="cycle-segment" data-value="${index}">第 ${index + 1} 段${index === metrics.currentIndex && !metrics.before && !metrics.after ? '<small>当前</small>' : ''}</button>`;}).join('')}</div>
    <div class="cycle-pager"><button class="link" data-action="cycle-segment" data-value="${selected - 1}" ${selected === 0 ? 'disabled' : ''}>← 上一段</button><small>第 ${selected + 1} / ${metrics.count} 段</small><button class="link" data-action="cycle-segment" data-value="${selected + 1}" ${selected === metrics.count - 1 ? 'disabled' : ''}>下一段 →</button></div>
    <section class="cycle-detail"><div class="between"><h3>${displayDate(segment.start)} — ${displayDate(segment.end)}</h3>${sourceTag(upcoming ? '未开始' : past ? '日期已过' : '当前段', upcoming ? 'blue' : past ? '' : 'green')}</div>
      ${past ? `<p>该时段有 ${daysWithRecords} 天留下记录，共 ${records.length} 条。未记录不能推断为未执行或恢复正常。</p>` : `<h4>当前已发布的每日安排</h4><div class="cycle-task-tags">${Object.keys(serviceTypes).filter(type => patient.plan[type]).map(type => sourceTag(serviceTypes[type])).join('')}</div><p>${escapeHtml(patient.plan.stage)} · V${patient.plan.version}。${upcoming ? '未来具体执行内容以届时医护确认的计划为准，不自动增加任务强度。' : '当天具体次数与注意事项在“今日”查看。'}</p>`}
      ${cycle.reviewEvery ? `<div class="cycle-checkpoint"><strong>${displayDate(segment.end)} · ${selected === metrics.count - 1 ? '周期末回访' : '计划回访节点'}</strong><p>${past ? '计划日期已过；本原型未记录该节点的回访结果，请与团队核实。' : '准备执行记录与问题，由医护核对后决定是否调整计划。'}</p></div>` : '<p>医护尚未设置回访间隔，本页不自动编排回访节点。</p>'}
      ${patient.appointment >= segment.start && patient.appointment <= segment.end ? `<div class="cycle-checkpoint"><strong>${displayDate(patient.appointment)} · 复诊安排</strong><p>由团队单独安排，不等同于周期回访，也不代表真实挂号成功。</p><button class="link" data-action="appointment-info" data-id="${patient.id}">查看复诊详情 →</button></div>` : ''}
      ${records.length ? `<div class="cycle-records"><h4>本段已提交记录</h4>${records.slice(0, 6).map(record => `<p><strong>${escapeHtml(record.title)}</strong><br>${displayDate(record.date)} · V${record.version} · ${escapeHtml(record.status)}</p>`).join('')}${records.length > 6 ? '<p>此处展示最近 6 条，其余记录可到“进度”查看。</p>' : ''}</div>` : '<p class="cycle-hint">本段暂无已提交记录。</p>'}</section>
    <div class="cycle-ending"><h3>周期结束后会怎样？</h3><p>团队结合记录和反馈，人工决定继续观察、调整计划或安排复诊。日期结束不等于康复完成，也不会自动关闭待处理事项。</p><div class="cycle-actions"><button class="btn small" data-action="view-plan" data-id="${patient.id}">阶段计划详情</button><button class="btn small" data-action="patient-page" data-value="progress">记录与周报</button></div></div>`;
}
