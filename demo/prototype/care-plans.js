const careTrainingSlots = {allDay:'全天总量',morning:'上午',afternoon:'下午',evening:'晚间'};
const careFoodSlots = {allDay:'全天总量',breakfast:'早餐',lunch:'午餐',dinner:'晚餐'};
const careSlots = {...careTrainingSlots,...careFoodSlots};
let careUi = {identity:'clinician',library:'actions',search:'',draft:null,pickerCategory:'all',currentSlot:'morning',currentMeal:'breakfast'};
const careClone = value => JSON.parse(JSON.stringify(value));
const careAdmin = () => ui.role === 'doctor' && careUi.identity === 'admin';
const careClinician = () => ui.role === 'doctor';
function ensureCareData() {
  if (!state.catalog) state.catalog = {
    actions:[
      {id:'action-1',name:'踝泵练习（示例）',category:'下肢',unit:'次',note:'由医护核对适用阶段、活动限制与动作要求；不适时停止并联系团队。',version:1,active:true},
      {id:'action-2',name:'坐位抬腿（示例）',category:'下肢',unit:'次',note:'需先由医护确认是否适用；本原型无真实动作示范或评估能力。',version:1,active:true},
      {id:'action-3',name:'肩部活动（示例）',category:'上肢',unit:'次',note:'活动范围与术后限制须由医护逐人确认，不能仅凭动作名称执行。',version:1,active:true}
    ],
    foods:[
      {id:'food-1',name:'米饭',category:'主食',unit:'克',note:'示例食材；份量口径由医护注明，不自动推算营养达标。',allergens:'配料与交叉接触需核对',version:1,active:true},
      {id:'food-2',name:'鸡蛋',category:'蛋白类',unit:'个',note:'烹调方式、质地与摄入限制须按个体情况确认。',allergens:'蛋类',version:1,active:true},
      {id:'food-3',name:'西兰花',category:'蔬菜',unit:'克',note:'食用方式与份量为计划示例，需医护确认。',allergens:'个体过敏史需核对',version:1,active:true},
      {id:'food-4',name:'牛奶',category:'奶类',unit:'毫升',note:'耐受情况与个体饮食限制需医护确认。',allergens:'乳类',version:1,active:true}
    ]
  };
  ensureFoodReferences();
  if (!Array.isArray(state.planTemplates)) state.planTemplates = [{
    id:'template-demo',name:'三餐记录 + 恢复观察',owner:currentClinician,version:1,active:true,
    content:{stage:'居家饮食与恢复观察',training:false,nutrition:true,survey:true,actions:[],foods:[
      {rowId:'demo-breakfast',catalogId:'food-2',snapshot:careClone(state.catalog.foods[1]),slot:'breakfast',amount:1},
      {rowId:'demo-lunch',catalogId:'food-1',snapshot:careClone(state.catalog.foods[0]),slot:'lunch',amount:100},
      {rowId:'demo-dinner',catalogId:'food-3',snapshot:careClone(state.catalog.foods[2]),slot:'dinner',amount:100}
    ],days:14,reviewEvery:7,kind:'观察周期',goal:'持续记录三餐和恢复反馈，在约定回访节点由医护核对。',note:'本模板仅为交互示例，不构成膳食建议。请医护核对过敏史、饮食限制与出院要求。'}
  }];
}
function careIdentityControl() {
  return `<label class="care-identity">演示权限<select id="care-identity"><option value="clinician" ${!careAdmin()?'selected':''}>普通医护 · 使用库 / 配计划</option><option value="admin" ${careAdmin()?'selected':''}>医护 + 内容管理员 · 维护库</option></select></label>`;
}
function careButton(action,label,id='',style='') {
  return `<button class="btn ${style}" data-action="care-${action}" data-id="${escapeHtml(id)}">${label}</button>`;
}
function renderCareLibrary() {
  ensureCareData();
  if(careUi.library==='products')return renderShopLibrary();
  const type=careUi.library, items=state.catalog[type].filter(item=>`${item.name} ${item.category}`.includes(careUi.search));
  return `<div class="page-heading between"><div><div class="eyebrow">CARE CONTENT LIBRARY</div><h1>照护内容库</h1><p>内容统一维护，医护按患者情况选择与组合。</p></div>${careAdmin()?careButton('catalog-new','＋ 新增'+(type==='actions'?'动作':'食材'),'','primary'):sourceTag('普通医护 · 只读使用','green')}</div>
    <div class="notice">${careAdmin()?'当前为医护 + 内容管理员演示身份，可新增、编辑和停用内容。':'普通医护可查看内容，并在计划编辑器中选用；只有内容管理员可以维护库。'} 库条目不是处方，选用仍需逐人核对。</div>
    <div class="care-toolbar"><div class="care-tabs">${['actions','foods'].map(key=>`<button class="btn ${key===type?'primary':''}" data-action="care-library-tab" data-id="${key}">${key==='actions'?'动作库':'食材库'} · ${state.catalog[key].length}</button>`).join('')}${careButton('library-tab','商品库','products')}</div><label class="care-search">搜索名称 / 分类<input id="care-search" value="${escapeHtml(careUi.search)}" placeholder="输入后回车搜索"></label></div>
    <div class="care-grid">${items.map(item=>`<article class="care-card"><div class="between"><span class="task-icon">${icon(type==='actions'?'training':'food')}</span>${sourceTag(item.active?'可选用':'已停用',item.active?'green':'')}</div><h3>${escapeHtml(item.name)}</h3><p>${escapeHtml(item.category)} · 单位：${escapeHtml(item.unit)} · V${item.version}</p><p>${escapeHtml(item.note)}</p>${type==='foods'?`<p>过敏 / 耐受提示：${escapeHtml(item.allergens)}</p>${foodReferenceCard(item,item.reference?.amount||1)}`:''}<div class="care-card-actions">${careButton('catalog-detail','查看详情',item.id)}${careAdmin()?careButton('catalog-edit','编辑',item.id)+careButton('catalog-toggle',item.active?'停用':'重新启用',item.id):''}</div></article>`).join('')||'<div class="empty">没有匹配内容，请更换搜索词。</div>'}</div>`;
}
function careCatalogForm(id) {
  if (!careAdmin()) return toast('仅内容管理员可以维护库。');
  const type=careUi.library, item=state.catalog[type].find(entry=>entry.id===id);
  modalContext={kind:'care-catalog',type,id:item?.id};
  showModal(item?'编辑库内容':'新增'+(type==='actions'?'动作':'食材'),'保存为库内容，不会自动修改已发布计划或历史模板。',
    field('名称',input('catalog-name',item?.name||'','text','maxlength="50"'))+field('分类',input('catalog-category',item?.category||'','text','maxlength="30"'))+
    field('计量单位',type==='actions'?'<input id="catalog-unit" value="次" readonly>':`<select id="catalog-unit">${['克','毫升','个','份'].map(unit=>`<option ${item?.unit===unit?'selected':''}>${unit}</option>`).join('')}</select>`)+
    field(type==='actions'?'执行说明 / 适用限制（医护核对）':'备餐说明 / 份量口径',textarea('catalog-note',item?.note||''))+
    (type==='foods'?field('过敏 / 耐受提示',input('catalog-allergens',item?.allergens||'','text','maxlength="100"'))+foodReferenceForm(item):'<div class="notice">当前仅支持按组 / 次配置。暂无真实示范视频与 AI 识别适配。</div>'),
    careButton('catalog-save','保存库内容','','primary'));
}
function careSaveCatalog() {
  if (!careAdmin() || modalContext.kind!=='care-catalog') return toast('没有维护权限。');
  const {type,id}=modalContext, existing=state.catalog[type].find(item=>item.id===id);
  const values={name:read('catalog-name'),category:read('catalog-category'),unit:read('catalog-unit'),note:read('catalog-note'),allergens:type==='foods'?read('catalog-allergens'):''};
  if (!values.name || !values.category || values.note.length<6 || (type==='foods'&&!values.allergens)) return formError('请填写名称、分类、至少 6 字说明，以及食材的过敏 / 耐受提示。');
  if (values.name.length>50 || values.category.length>30 || values.note.length>1000 || values.allergens.length>100) return formError('名称最多 50 字，分类 30 字，说明 1000 字，过敏提示 100 字。');
  if (state.catalog[type].some(item=>item.id!==id&&item.name===values.name)) return formError('库中已有同名内容，请编辑原条目或使用不同名称。');
  if (!(type==='actions'?['次']:['克','毫升','个','份']).includes(values.unit)) return formError('请选择有效单位。');
  if(type==='foods'){values.reference=readFoodReference();const error=foodReferenceValidation(values.reference);if(error)return formError(error);}
  if (existing) Object.assign(existing,values,{version:existing.version+1});
  else state.catalog[type].push({...values,id:uuid(type),version:1,active:true});
  addEvent(`内容管理员 · ${existing?'更新':'新增'}库内容：${values.name}`);closeModal();persist();toast('库已保存；既有计划和模板快照保持不变');
}
function renderCareTemplates() {
  ensureCareData();
  return `<div class="page-heading between"><div><div class="eyebrow">REUSABLE CARE PLANS</div><h1>常用计划模板</h1><p>保存常用搭配，套用后仍需按患者情况确认。</p></div>${careButton('template-new','＋ 新建模板','','primary')}</div><div class="notice">模板保存相对天数与回访间隔，不包含患者资料、日历日期或执行记录。当前维护范围：${escapeHtml(currentClinician)}的个人常用模板。</div><div class="care-grid">${state.planTemplates.filter(item=>item.active).map(item=>`<article class="care-card"><div class="between">${sourceTag('个人模板','green')}<small>V${item.version}</small></div><h3>${escapeHtml(item.name)}</h3><p>${item.content.days} 天 · 每 ${item.content.reviewEvery} 天回访</p><p>${item.content.training?item.content.actions.length:0} 项训练 · ${item.content.nutrition?item.content.foods.length:0} 项食材 · ${item.content.survey?'每日随访':'无每日随访'}</p><p>${escapeHtml(item.content.goal)}</p><div class="care-card-actions">${careButton('template-assign','给患者设置',item.id,'primary')}${careButton('template-edit','查看 / 编辑',item.id)}${careButton('template-copy','复制',item.id)}${careButton('template-archive','归档',item.id)}</div></article>`).join('')||'<div class="empty">暂无常用模板，可以新建或从患者计划中保存。</div>'}</div>`;
}
function openCareBuilder(pid,returnCase=null,templateId=null) {
  if (!careClinician()) return toast('需由医护设置计划。');
  ensureCareData();
  if (careUi.draft) return showModal('有一份尚未保存的草稿','请先继续或放弃当前草稿。','<p>不会覆盖正在编辑的内容，也不会改变患者当前计划。</p>',careButton('builder-resume','继续编辑','','primary')+careButton('builder-discard','放弃草稿'));
  const patient=patientById(pid), template=state.planTemplates.find(item=>item.id===templateId),published=patientHasPublishedPlan(patient);
  const plan=published?careClone(patient.plan):{stage:'居家恢复计划',training:false,nutrition:false,survey:true,note:'请按医护确认的计划执行；不适时停止当前任务并联系团队。'};
  const cycle=published?cycleFor(patient):{start:TODAY,kind:patient?.path==='检查后随访'?'观察周期':'康复计划',reviewEvery:7,goal:'持续记录居家执行和不适反馈，在回访节点由医护核对。'};
  careUi.draft={...plan,pid:patient?.id||null,returnCase,templateId:patient?null:templateId,templateName:template?.name||'',cycle:careClone(cycle),end:patient?.plan.end||dayDate(dateDay(TODAY)+13),actions:careClone(plan.actions||[]),foods:careClone(plan.foods||[]),reason:'',confirmed:false,baseVersion:patient?.plan.version||0,legacy:published&&!plan.actions};
  careUi.currentSlot='morning';careUi.currentMeal='breakfast';
  if (template) careApplyTemplate(template);
  closeModal();ui.page='builder';render();
}
function careApplyTemplate(template) {
  const draft=careUi.draft,content=careClone(template.content), start=draft.cycle.start;
  Object.assign(draft,{stage:content.stage,training:content.training,nutrition:content.nutrition,survey:content.survey,note:content.note,actions:content.actions.map(row=>({...row,rowId:uuid('task')})),foods:content.foods.map(row=>({...row,rowId:uuid('food-row')})),end:dayDate(dateDay(start)+content.days-1),cycle:{start,kind:content.kind,reviewEvery:content.reviewEvery,goal:content.goal},confirmed:false,legacy:false});
}
function careSelect(id,values,current) {
  return `<select id="${id}">${Object.entries(values).map(([value,label])=>`<option value="${value}" ${String(current)===value?'selected':''}>${escapeHtml(label)}</option>`).join('')}</select>`;
}
function careRows(type) {
  const draft=careUi.draft, slots=type==='actions'?careTrainingSlots:careFoodSlots,selected=type==='actions'?careUi.currentSlot:careUi.currentMeal;
  const toolbar=`<div class="care-section-toolbar"><div class="care-slot-picker" aria-label="${type==='actions'?'执行时段':'餐次'}">${Object.entries(slots).map(([slot,label])=>`<button class="btn ${slot===selected?'primary':''}" data-action="care-slot" data-id="${type}:${slot}" aria-pressed="${slot===selected}">${label}<span>${draft[type].filter(row=>row.slot===slot).length}</span></button>`).join('')}</div>${careButton('pick',type==='actions'?'＋ 从动作库添加':'＋ 从食材库添加',type)}</div>`;
  const rows=draft[type].filter(row=>row.slot===selected).map(row=>{
    const library=state.catalog[type].find(item=>item.id===row.catalogId), outdated=library?.version!==row.snapshot.version;
    return `<article class="care-plan-row compact"><div class="care-row-line ${type}"><div class="care-row-main"><h3>${escapeHtml(row.snapshot.name)}</h3><small>${escapeHtml(row.snapshot.category)} · 库 V${row.snapshot.version}</small></div><label class="care-inline-field"><span class="care-sr-only">${escapeHtml(row.snapshot.name)}时段 / 餐次</span>${careSelect(`row-slot-${row.rowId}`,slots,row.slot)}</label>${type==='actions'?`<label class="care-inline-field"><span class="care-sr-only">组数</span>${input(`row-sets-${row.rowId}`,row.sets,'number','min="1" max="5"')}<span class="care-mobile-label">组</span></label><label class="care-inline-field"><span class="care-sr-only">每组次数</span>${input(`row-reps-${row.rowId}`,row.reps,'number','min="1" max="30"')}<span class="care-mobile-label">次 / 组</span></label>`:`<label class="care-inline-field"><span class="care-sr-only">计划份量</span>${input(`row-amount-${row.rowId}`,row.amount,'number','min="0.01" max="3000" step="0.01"')}<span>${escapeHtml(row.snapshot.unit)}</span></label>`}${careButton('row-remove','移除',`${type}:${row.rowId}`)}</div>${type==='foods'?foodRowReference(row,true):''}<details class="care-row-note"><summary>查看适用说明${type==='foods'?'与过敏提示':''}</summary><p>${escapeHtml(row.snapshot.note)}${type==='foods'?` · 过敏 / 耐受：${escapeHtml(row.snapshot.allergens)}`:''}</p></details>${!library?.active?'<div class="notice red">原条目已停用，请移除并选择可用内容。</div>':outdated?`<div class="notice">库有新版本；请重新选用并核对。${careButton('row-refresh','采用当前库版本',`${type}:${row.rowId}`)}</div>`:''}</article>`;
  }).join('');
  return toolbar+(rows?`<div class="care-row-heading care-row-line ${type}" aria-hidden="true"><span>${type==='actions'?'动作':'食材'}</span><span>${type==='actions'?'时段':'餐次'}</span>${type==='actions'?'<span>组数</span><span>次数 / 组</span>':'<span>计划份量</span>'}<span>操作</span></div>${rows}`:`<div class="care-placeholder">${type==='actions'?`${careSlots[selected]}未安排，可留空；如需安排，请从动作库添加。`:`${careSlots[selected]}未配置，可留空；如需安排，请从食材库添加。`}</div>`);
}
function renderCareBuilder() {
  const draft=careUi.draft;if(!draft)return renderPlans();
  const patient=patientById(draft.pid), days=dateDay(draft.end)-dateDay(draft.cycle.start)+1;
  return `<div class="page-heading between"><div><div class="eyebrow">CARE PLAN COMPOSER</div><h1>${patient?'为'+escapeHtml(patient.name)+'设置计划':'编辑常用模板'}</h1><p>${patient?`${escapeHtml(patient.procedure)} · ${patientHasPublishedPlan(patient)?`当前 V${patient.plan.version}，发布后生成新版本`:'当前尚未建立计划，首次发布后生成 V1'}`:'可复用的内容搭配，不会直接发布给患者'}</p></div>${careButton('builder-leave','返回列表')}</div><div class="care-stepbar"><span>01 选用内容 / 模板</span><span>02 设置执行与周期</span><span>03 医护核对${patient?'发布':'保存'}</span></div>
    <div id="form-error" class="validation-error" role="alert"></div><div class="care-composer"><div>
    <section class="care-section"><h2>基本信息与快捷套用</h2><div class="care-toolbar"><select id="care-template-choice" aria-label="选择常用模板"><option value="">选择常用模板…</option>${state.planTemplates.filter(item=>item.active).map(item=>`<option value="${item.id}">${escapeHtml(item.name)} · ${item.content.days}天</option>`).join('')}</select>${careButton('template-apply','套用到草稿')}</div><p class="care-caption">套用会替换草稿内容，并按当前周期起点重新计算结束日期；不会直接发布。</p>${!patient?field('模板名称',input('care-template-name',draft.templateName,'text','maxlength="50"')):''}${field('阶段名称',input('care-stage',draft.stage,'text','maxlength="50"'))}${draft.legacy?'<div class="notice">旧版计划只有总组次 / 营养目标，未关联具体库内容。请重新选用，不自动猜测动作或食材。</div>':''}<div class="flex care-services">${Object.entries(serviceTypes).map(([type,label])=>`<label class="checkbox"><input id="care-${type}" type="checkbox" ${draft[type]?'checked':''}>${label}</label>`).join('')}</div></section>
    ${draft.training?`<section class="care-section"><h2>康复训练 · 每日执行</h2><p class="care-caption">按需选择全天总量或具体时段，再添加动作；只安排上午也可以，其他时段可留空。全天总量不限定时段，组次为一天合计，无需在上午 / 下午 / 晚间重复添加；不同条目的训练量会累加，均由医护核对。</p>${careRows('actions')}</section>`:''}
    ${draft.nutrition?`<section class="care-section"><h2>饮食计划 · 全天 / 按餐搭配</h2><p class="care-caption">按需选择全天总量或具体餐次，无需配齐三餐。全天份量为一天合计，不是每餐都吃这些量；未配置某餐不代表禁食。同一食材请选择全天量或按餐分配，不要重复配置。份量与能量仅为参考，不代填患者实食记录。</p>${careRows('foods')}</section>`:''}
    <section class="care-section"><h2>服务周期与回访</h2>${patient?cycleEditor({plan:draft,versions:patient.versions}):field('周期类型',careSelect('cycle-kind',{'康复计划':'康复计划','观察周期':'观察周期'},draft.cycle.kind))+field('每隔多少天回访',input('cycle-interval',draft.cycle.reviewEvery,'number','min="1" max="90"'))+field('周期目标',textarea('cycle-goal',draft.cycle.goal))}${field('持续天数（含首尾日）',input('care-days',days,'number','min="1" max="366"'),patient?'修改天数会同步结束日期；修改起止日期也会更新天数。':'给患者套用时再确定起点；模板不保存具体日期。')}</section>
    <section class="care-section"><h2>注意事项与确认</h2>${field('患者可见的注意事项',textarea('care-note',draft.note))}${patient?field(patientHasPublishedPlan(patient)?'本次调整原因':'首次计划制定说明',textarea('care-reason',draft.reason)):'<div class="notice">模板仅为医护效率工具，不意味着适用于任意患者。</div>'}<label class="checkbox"><input id="care-confirm" type="checkbox" ${draft.confirmed?'checked':''}><span>已核对适用阶段、动作限制、过敏 / 饮食要求和周期；本原型不用于真实诊疗。</span></label></section></div>
    <aside class="care-preview"><div class="eyebrow">PLAN PREVIEW</div><h2>本次安排</h2><strong class="care-days">${Number.isFinite(days)?days:'—'}<small>天</small></strong><p>${patient?`${displayDate(draft.cycle.start)} — ${displayDate(draft.end)}`:'相对周期 · 不绑定患者日期'}</p><p>每 ${draft.cycle.reviewEvery} 天回访</p><hr class="rule">${careContentSummary(draft)}<div class="notice">${patient?'确认前仅为草稿。发布只更新计划，不会关闭高风险事项。':'保存后进入个人常用模板库，套用时仍需重新确认。'}</div>${patient?careButton('publish',patientHasPublishedPlan(patient)?'确认发布新版本':'确认发布首个计划 · V1','','primary block')+careButton('save-as-template','另存为常用模板','','block'):careButton('template-save','保存模板','','primary block')}<p class="care-caption">草稿仅在本次页面会话保留，刷新会丢失；已发布内容保存在本地浏览器。</p></aside></div>`;
}
function careContentSummary(plan) {
  return `<div class="care-content-summary">${plan.training?`<h3>每日训练 · ${plan.actions.length} 项</h3>${plan.actions.map(row=>`<p>${escapeHtml(careSlots[row.slot])} · ${escapeHtml(row.snapshot.name)}<br>${row.sets} 组 × ${row.reps} 次</p>`).join('')}`:''}${plan.nutrition?`<h3>每日饮食计划</h3><p>全天量为一天合计，不是每餐份量；各分组列出的不同食材合并计入每日参考。</p>${Object.keys(careFoodSlots).filter(slot=>plan.foods.some(row=>row.slot===slot)).map(slot=>{const rows=plan.foods.filter(row=>row.slot===slot);return `<h4>${careSlots[slot]}</h4>${rows.map(row=>`<p>${escapeHtml(row.snapshot.name)} ${foodNumber(row.amount)} ${escapeHtml(row.snapshot.unit)}<br><small>${foodEquivalent(row.snapshot,row.amount)} · ${foodEnergyText(foodEstimate(row.snapshot,row.amount).kcal)}</small></p>`).join('')||'<p>未配置</p>'}${foodEnergyTotal(rows,slot==='allDay'?'全天未分餐食材':'本餐已配置')}`;}).join('')}${foodEnergyTotal(plan.foods,'每日已配置食材')}`:''}${plan.survey?'<p>恢复随访 · 每日一次</p>':''}</div>`;
}
function careReadDraft() {
  const draft=careUi.draft;if(!draft||!document.getElementById('care-stage'))return;
  const duration=Number(read('care-days'));
  draft.invalidDuration=!Number.isInteger(duration)||duration<1||duration>366;
  Object.assign(draft,{stage:read('care-stage'),note:read('care-note'),confirmed:checked('care-confirm'),training:checked('care-training'),nutrition:checked('care-nutrition'),survey:checked('care-survey')});
  if(draft.pid){draft.reason=read('care-reason');draft.cycle=readCycle();draft.end=read('plan-end');}
  else {draft.templateName=read('care-template-name');Object.assign(draft.cycle,{kind:read('cycle-kind'),reviewEvery:Number(read('cycle-interval')),goal:read('cycle-goal')});const days=Number(read('care-days'));draft.end=Number.isInteger(days)&&days>=1&&days<=366?dayDate(dateDay(draft.cycle.start)+days-1):'';}
  for(const type of ['actions','foods']) for(const row of draft[type]) if(document.getElementById(`row-slot-${row.rowId}`)) {
    row.slot=read(`row-slot-${row.rowId}`);
    if(type==='actions'){row.sets=Number(read(`row-sets-${row.rowId}`));row.reps=Number(read(`row-reps-${row.rowId}`));}else row.amount=Number(read(`row-amount-${row.rowId}`));
  }
}
function careValidation(draft) {
  if(draft.invalidDuration)return '持续天数须为 1–366 的整数。';
  if(!draft.stage||draft.stage.length>50||draft.note.length<6||draft.note.length>1000)return '请填写阶段名称（最多 50 字）与 6–1000 字的注意事项。';
  const error=cycleValidation(draft.cycle,draft.end);if(error)return error;
  if(draft.pid&&draft.end<TODAY)return '患者计划结束日期不能早于当前演示日。';
  if(draft.pid&&draft.cycle.start>TODAY)return '本版发布即时生效，周期起点不能晚于当前演示日；未来生效需另行设计。';
  if(!draft.training&&!draft.nutrition&&!draft.survey)return '请至少启用一种服务。';
  for(const type of ['actions','foods']) {
    if(!(type==='actions'?draft.training:draft.nutrition))continue;
    if(!draft[type].length)return `请从${type==='actions'?'动作':'食材'}库选择至少一项内容。`;
    for(const row of draft[type]) {
      const item=state.catalog[type].find(entry=>entry.id===row.catalogId);
      if(!item?.active||item.version!==row.snapshot.version)return `「${row.snapshot.name}」已停用或有新版本，请重新选用 / 更新快照后核对。`;
      if(!(Object.keys(type==='actions'?careTrainingSlots:careFoodSlots)).includes(row.slot))return '请选择有效时段 / 餐次。';
      if(type==='actions'&&(!Number.isInteger(row.sets)||row.sets<1||row.sets>5||!Number.isInteger(row.reps)||row.reps<1||row.reps>30))return '每项训练须为 1–5 组，每组 1–30 次（原型范围，不是医疗标准）。';
      if(type==='foods'&&(!Number.isFinite(row.amount)||row.amount<=0||row.amount>3000))return '食材份量须大于 0 且不超过 3000（按所选单位，原型范围）。';
    }
  }
  if(draft.nutrition){
    const overlap=draft.foods.find(row=>row.slot==='allDay'&&draft.foods.some(other=>other.catalogId===row.catalogId&&other.slot!=='allDay'));
    if(overlap)return `「${overlap.snapshot.name}」同时设置了全天量与分餐量，可能重复计算。请保留全天量，或移除全天条目后按餐分配。`;
  }
  return '';
}
function careTemplateContent(draft) {
  return careClone({stage:draft.stage,training:draft.training,nutrition:draft.nutrition,survey:draft.survey,actions:draft.training?draft.actions:[],foods:draft.nutrition?draft.foods:[],days:dateDay(draft.end)-dateDay(draft.cycle.start)+1,reviewEvery:draft.cycle.reviewEvery,kind:draft.cycle.kind,goal:draft.cycle.goal,note:draft.note});
}
function careSaveTemplate(asNew=false) {
  if(!careClinician()||!careUi.draft)return;
  careReadDraft();const draft=careUi.draft,error=careValidation(draft);if(error)return formError(error);
  if(!draft.confirmed)return formError('请先完成医护核对。');
  if(asNew){modalContext={kind:'care-template-name'};return showModal('另存为个人常用模板','只保存内容搭配和相对周期，不保存患者信息。',field('模板名称',input('template-save-name',draft.stage,'text','maxlength="50"')),careButton('template-save-named','保存模板','','primary'));}
  careCommitTemplate(draft.templateName,draft.templateId);
}
function careCommitTemplate(name,id=null) {
  if(!careClinician()||!careUi.draft)return;
  const draft=careUi.draft,error=careValidation(draft);if(error)return formError(error);
  if(!draft.confirmed)return formError('请完成医护核对。');
  if(!name||name.length>50)return formError('模板名称须为 1–50 字。');
  const existing=state.planTemplates.find(item=>item.id===id);
  if(existing&&existing.owner!==currentClinician)return formError('仅可维护自己的常用模板。');
  if(state.planTemplates.some(item=>item.active&&item.owner===currentClinician&&item.id!==id&&item.name===name))return formError('已有同名模板，请使用其他名称。');
  const saved={id:id||uuid('template'),name,owner:currentClinician,active:true,version:(existing?.version||0)+1,content:careTemplateContent(draft)};
  if(existing)Object.assign(existing,saved);else state.planTemplates.push(saved);
  if(!draft.pid){careUi.draft=null;ui.page='templates';}closeModal();persist();toast('已保存常用模板，未改变患者当前计划');
}
function carePublish() {
  if(!careClinician()||!careUi.draft?.pid)return;
  careReadDraft();const draft=careUi.draft, patient=patientById(draft.pid),error=careValidation(draft);
  if(error)return formError(error);
  if((patient.plan.version||0)!==draft.baseVersion)return formError('患者计划已在其他窗口更新，请放弃草稿并重新打开后核对。');
  if(draft.reason.length<6||draft.reason.length>500)return formError('请填写 6–500 字的计划制定或调整说明。');
  if(!draft.confirmed)return formError('请勾选医护核对后发布。');
  const firstPublish=!patientHasPublishedPlan(patient);
  const next={...careClone(patient.plan),status:'已发布',version:(patient.plan.version||0)+1,stage:draft.stage,start:TODAY,end:draft.end,cycle:careClone(draft.cycle),training:draft.training,nutrition:draft.nutrition,survey:draft.survey,actions:careClone(draft.training?draft.actions:[]),foods:careClone(draft.nutrition?draft.foods:[]),note:draft.note,reason:draft.reason,author:currentClinician};
  const returnCase=draft.returnCase;
  patient.plan=next;patient.versions.push(careClone(next));
  addMessage(patient.id,`阶段计划已${firstPublish?'建立':'更新'}为 V${next.version}`,`周期：${displayDate(next.cycle.start)} 至 ${displayDate(next.end)}，每 ${next.cycle.reviewEvery} 天回访。${next.reason}。请查看具体动作、餐次搭配和注意事项：${next.note}。${firstPublish?'':'如今日已有旧版本训练记录，新版训练从明日起执行，今日不重复补做。'}`);
  addEvent(`${patient.name} · 医护发布内容库组合计划 V${next.version}`);
  const item=state.cases.find(entry=>entry.id===returnCase);if(item)item.audit.push({time:`10-08 ${timeLabel()}`,text:`${currentClinician}发布计划 V${next.version}：${next.reason}。事项未自动关闭。`});
  careUi.draft=null;ui.page='plans';persist();if(returnCase)caseDetail(returnCase);toast('新计划已发布，患者端同步；高风险事项保持原状态');
}
function carePicker(type,category='all') {
  if(!['actions','foods'].includes(type)||!careUi.draft)return;
  careUi.pickerCategory=category;
  careReadDraft();modalContext={kind:'care-picker',type};
  const categories=[...new Set(state.catalog[type].filter(item=>item.active).map(item=>item.category))],labels={'下肢':'下肢动作','上肢':'上肢动作','蛋白类':'肉类 / 蛋白'},slot=type==='actions'?careUi.currentSlot:careUi.currentMeal;
  showModal(type==='actions'?'从动作库选择':'从食材库选择',`正在配置${careSlots[slot]} · 可连续添加，返回计划后核对${type==='actions'?'组数与次数':'份量'}。`,`<div class="care-picker-toolbar">${['all',...categories].map(value=>`<button class="btn ${value===category?'primary':''}" data-action="care-pick-category" data-id="${escapeHtml(value)}" aria-pressed="${value===category}">${value==='all'?'全部':escapeHtml(labels[value]||value)}</button>`).join('')}</div><div class="care-picker">${state.catalog[type].filter(item=>item.active&&(category==='all'||item.category===category)).map(item=>`<article class="care-picker-item"><div class="care-picker-item-head"><div><h3>${escapeHtml(item.name)}</h3><small>${escapeHtml(labels[item.category]||item.category)} · V${item.version}${careUi.draft[type].some(row=>row.catalogId===item.id&&row.slot===slot)?` · ${careSlots[slot]}已添加 ${careUi.draft[type].filter(row=>row.catalogId===item.id&&row.slot===slot).length} 项`:''}</small></div>${careButton('pick-add','＋ 添加',item.id,'primary')}</div>${type==='foods'?`<p class="care-picker-reference">${foodNumber(item.reference?.amount||1)} ${escapeHtml(item.unit)} · ${foodEquivalent(item,item.reference?.amount||1)} · ${foodEnergyText(foodEstimate(item,item.reference?.amount||1).kcal)}</p>`:''}<p>${escapeHtml(item.note)}${type==='foods'?' · 过敏 / 耐受：'+escapeHtml(item.allergens):''}</p></article>`).join('')||'<div class="empty">此分类暂无可用内容，请联系管理员新增。</div>'}</div>`,careButton('pick-done','完成选择，返回计划','','primary'));
}
function careDraftInput(event) {
  if(ui.page!=='builder'||!careUi.draft||!event.target.closest('.care-composer'))return;
  const target=event.target,draft=careUi.draft;
  if(target.type==='checkbox'||target.tagName==='SELECT')return;
  if(target.id==='care-days'&&draft.pid){
    const days=Number(target.value),start=dateDay(read('cycle-start'));
    if(Number.isInteger(days)&&days>=1&&days<=366&&Number.isFinite(start))document.getElementById('plan-end').value=dayDate(start+days-1);
  }
  if(['cycle-start','plan-end'].includes(target.id)){
    const days=dateDay(read('plan-end'))-dateDay(read('cycle-start'))+1;
    document.getElementById('care-days').value=Number.isFinite(days)?days:'';
  }
  careReadDraft();draft.confirmed=false;document.getElementById('care-confirm').checked=false;
  refreshFoodRowReferences();
  const preview=document.querySelector('.care-content-summary');if(preview)preview.outerHTML=careContentSummary(draft);
  const days=dateDay(draft.end)-dateDay(draft.cycle.start)+1;
  document.querySelector('.care-days').innerHTML=`${Number.isFinite(days)?days:'—'}<small>天</small>`;
  const paragraphs=document.querySelectorAll('.care-preview>p');
  if(draft.pid)paragraphs[0].textContent=`${displayDate(draft.cycle.start)} — ${displayDate(draft.end)}`;
  paragraphs[1].textContent=`每 ${draft.cycle.reviewEvery} 天回访`;
}
document.addEventListener('input',careDraftInput);
document.addEventListener('keydown',event=>{
  if(event.target.id==='care-search'&&event.key==='Enter'){careUi.search=event.target.value.trim();render();}
});
document.addEventListener('change',event=>{
  if(event.target.id==='care-identity'){careReadDraft();careUi.identity=event.target.value==='admin'?'admin':'clinician';closeModal();render();return;}
  if(event.target.id==='care-search'){careUi.search=event.target.value.trim();render();return;}
  if(ui.page!=='builder'||!careUi.draft||!event.target.closest('.care-composer'))return;
  const draft=careUi.draft;
  if(event.target.id==='care-days'&&draft.pid){const days=Number(event.target.value);if(!Number.isInteger(days)||days<1||days>366)return formError('持续天数须为 1–366 的整数。');const start=read('cycle-start');if(!Number.isFinite(dateDay(start)))return formError('请先填写有效的开始日期。');document.getElementById('plan-end').value=dayDate(dateDay(start)+days-1);}
  careReadDraft();if(event.target.id!=='care-confirm')draft.confirmed=false;
  refreshFoodRowReferences();
  if(['care-training','care-nutrition','care-survey'].includes(event.target.id)||event.target.id.startsWith('row-slot-'))render();
  else {const preview=document.querySelector('.care-content-summary');if(preview)preview.outerHTML=careContentSummary(draft);if(event.target.id!=='care-confirm')document.getElementById('care-confirm').checked=false;}
});
document.addEventListener('click',event=>{
  const button=event.target.closest('[data-action^="care-"]');if(!button||button.disabled)return;
  event.stopImmediatePropagation();const action=button.dataset.action.slice(5),id=button.dataset.id;
  if(!careClinician())return toast('请在医护工作台操作。');
  ensureCareData();
  if(action==='library-tab'){careUi.library=id;careUi.search='';render();}
  if(action==='catalog-new'||action==='catalog-edit')careCatalogForm(id);
  if(action==='catalog-save')careSaveCatalog();
  if(action==='catalog-detail'){const item=state.catalog[careUi.library].find(entry=>entry.id===id);if(item)showModal(escapeHtml(item.name),`库内容 V${item.version} · ${item.active?'可选用':'已停用'}`,`<p>${escapeHtml(item.note)}</p><p>分类：${escapeHtml(item.category)} · 单位：${escapeHtml(item.unit)}</p>${item.allergens?`<p>过敏 / 耐受：${escapeHtml(item.allergens)}</p>`:''}${careUi.library==='foods'?foodReferenceCard(item,item.reference?.amount||1):''}<div class="notice">医护需核对个体适用性。编辑库不会改变已发布计划的快照。</div>`);}
  if(action==='catalog-toggle'){if(!careAdmin())return toast('仅内容管理员可以维护库。');const item=state.catalog[careUi.library].find(entry=>entry.id===id);if(item){item.active=!item.active;addEvent(`内容管理员 · ${item.active?'启用':'停用'} ${item.name}`);persist();toast('仅影响后续选用；已发布患者计划不自动变化，请人工核对');}}
  if(action==='template-new')openCareBuilder(null);
  if(action==='template-edit')openCareBuilder(null,null,id);
  if(action==='template-copy'){const item=state.planTemplates.find(entry=>entry.id===id);if(item){state.planTemplates.push({...careClone(item),id:uuid('template'),name:item.name+' · 副本',version:1,owner:currentClinician});persist();}}
  if(action==='template-archive'){const item=state.planTemplates.find(entry=>entry.id===id);if(item?.owner===currentClinician){item.active=false;persist();toast('模板已归档，患者已发布计划不受影响');}}
  if(action==='template-assign'){modalContext={kind:'care-assign',templateId:id};showModal('选择患者','套用后进入草稿，核对内容与周期再发布。',field('患者',`<select id="care-assign-patient">${state.patients.map(patient=>`<option value="${patient.id}">${escapeHtml(patient.name)} · ${escapeHtml(patient.procedure)}</option>`).join('')}</select>`),careButton('template-assign-confirm','进入患者计划','','primary'));}
  if(action==='template-assign-confirm')openCareBuilder(read('care-assign-patient'),null,modalContext.templateId);
  if(action==='template-apply'){careReadDraft();const template=state.planTemplates.find(item=>item.id===read('care-template-choice'));if(!template)return toast('请先选择模板。');modalContext={kind:'care-template-apply',templateId:template.id};showModal('确认替换当前草稿？','患者当前计划不受影响',`<p>将采用「${escapeHtml(template.name)}」的内容、${template.content.days} 天时长和回访间隔，并保留当前周期起点。原草稿的搭配将被替换。</p>`,careButton('template-apply-confirm','确认套用','','primary'));}
  if(action==='template-apply-confirm'){const template=state.planTemplates.find(item=>item.id===modalContext.templateId);if(template&&careUi.draft){careApplyTemplate(template);closeModal();render();}}
  if(action==='slot'&&careUi.draft){careReadDraft();const [type,slot]=id.split(':');if((Object.keys(type==='actions'?careTrainingSlots:careFoodSlots)).includes(slot)){careUi[type==='actions'?'currentSlot':'currentMeal']=slot;render();}}
  if(action==='pick')carePicker(id);
  if(action==='pick-category'&&modalContext.kind==='care-picker')carePicker(modalContext.type,id);
  if(action==='pick-done'){closeModal();render();}
  if(action==='food-portion'&&careUi.draft){const [rowId,scale]=id.split(':');const row=careUi.draft.foods.find(entry=>entry.rowId===rowId);if(row?.snapshot.reference&&[0.5,1,1.5,2].includes(Number(scale))){const amount=Math.round(row.snapshot.reference.amount*Number(scale)*100)/100;if(amount<=0||amount>3000)return toast('换算超出份量输入范围，请手动核对。');const target=document.getElementById(`row-amount-${rowId}`);target.value=amount;careDraftInput({target});}}
  if(action==='pick-add'&&careUi.draft){const type=modalContext.type,item=state.catalog[type]?.find(entry=>entry.id===id&&entry.active);if(item){careReadDraft();careUi.draft[type].push({rowId:uuid('row'),catalogId:id,snapshot:careClone(item),slot:type==='actions'?careUi.currentSlot:careUi.currentMeal,...(type==='actions'?{sets:1,reps:1}:{amount:item.reference?.amount||1})});careUi.draft.confirmed=false;render();carePicker(type,careUi.pickerCategory);toast('已添加，可继续选择或返回计划设置参数');}}
  if(['row-remove','row-refresh'].includes(action)&&careUi.draft){careReadDraft();const [type,rowId]=id.split(':');const row=careUi.draft[type].find(item=>item.rowId===rowId);if(action==='row-remove')careUi.draft[type]=careUi.draft[type].filter(item=>item.rowId!==rowId);else{const item=state.catalog[type].find(item=>item.id===row.catalogId&&item.active);if(item)row.snapshot=careClone(item);}careUi.draft.confirmed=false;render();}
  if(action==='publish')carePublish();
  if(action==='template-save')careSaveTemplate();
  if(action==='save-as-template')careSaveTemplate(true);
  if(action==='template-save-named')careCommitTemplate(read('template-save-name'));
  if(action==='builder-leave'){careReadDraft();showModal('离开计划编辑？','尚未发布的草稿不会改变患者计划。','<p>可保留草稿稍后继续，也可放弃本次编辑。</p>',careButton('builder-keep','保留草稿并返回')+careButton('builder-discard','放弃草稿'));}
  if(action==='builder-resume'){closeModal();ui.page='builder';render();}
  if(action==='builder-keep'||action==='builder-discard'){const returnCase=careUi.draft?.returnCase;ui.page=careUi.draft?.pid?'plans':'templates';if(action==='builder-discard')careUi.draft=null;closeModal();render();if(returnCase)caseDetail(returnCase);}
},true);
function configuredTrainingDeferred(patient) {
  return Array.isArray(patient.plan.actions)&&patientRecords(patient.id).some(record=>record.type==='training'&&record.date===TODAY&&record.version!==patient.plan.version);
}
function carePatientMeal(slot,patient=activePatient()) {
  const foods=patient.plan.foods?.filter(row=>row.slot===slot)||[],allDay=slot==='allDay';
  if(!foods.length)return !allDay&&Array.isArray(patient.plan.foods)?'<p class="task-hint">本餐未单独配置食材，不代表禁食；请结合全天计划与医护说明，记录实际用餐。</p>':'';
  return `<div class="patient-meal-plan"><strong>医护设置的${careSlots[slot]}${allDay?'':'搭配'}</strong>${allDay?'<p>以下为一天合计，不是每餐份量，不需要每餐重复吃；实际摄入请分别记在早餐、午餐、晚餐中。</p>':''}${foods.map(row=>`<section class="patient-food-item"><h4>${escapeHtml(row.snapshot.name)} · ${foodNumber(row.amount)} ${escapeHtml(row.snapshot.unit)}</h4>${foodReferenceCard(row.snapshot,row.amount)}<p>${escapeHtml(row.snapshot.note)}<br>过敏 / 耐受：${escapeHtml(row.snapshot.allergens)}</p></section>`).join('')}${foodEnergyTotal(foods,allDay?'全天未分餐食材':'本餐已配置')}<small>这是计划参考，不是已吃下的份量；实际用餐仍需单独记录。</small></div>`;
}
