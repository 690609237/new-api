function taskSlots(patient,type){
  if(type==='training'&&Array.isArray(patient.plan.actions))return patient.plan.actions.map(row=>({id:row.rowId,label:row.snapshot.name,time:row.slot==='allDay'?'全天总量（不限定时段）':careSlots[row.slot],sets:row.sets,reps:row.reps,note:row.snapshot.note}));
  if(type==='nutrition')return [{id:'breakfast',label:'早餐',time:'早间'},{id:'lunch',label:'午餐',time:'午间'},{id:'dinner',label:'晚餐',time:'晚间'}];
  if(type==='training')return [{id:'morning',label:'上午训练',time:'上午',sets:Math.ceil(patient.plan.sets/2)},{id:'afternoon',label:'下午训练',time:'下午',sets:Math.floor(patient.plan.sets/2)}].filter(slot=>slot.sets>0);
  return [{id:'daily',label:'今日恢复随访',time:'每日一次'}];
}
function slotRecord(patient,type,slot){return patientRecords(patient.id).find(record=>record.type===type&&record.date===TODAY&&record.slot===slot&&(type!=='training'||!Array.isArray(patient.plan.actions)||record.version===patient.plan.version));}
function legacyTrainingDone(patient){if(Array.isArray(patient.plan.actions))return false;return patientRecords(patient.id).some(record=>record.type==='training'&&record.date===TODAY&&!record.slot&&record.status==='已完成');}
function slotDone(patient,type,slot){return slotRecord(patient,type,slot)?.status==='已完成';}
function taskSlot(){return taskSlots(activePatient(),ui.task.type).find(slot=>slot.id===ui.task.slot);}
function openPatientTask(type){
  if(!enabledTasks(activePatient()).includes(type))return toast('当前计划未安排此任务，请与团队确认。');
  closeModal();ui.patientPage='task';ui.task={type,slot:null,body:null};render();
}
function startTaskSlot(slotId){
  const patient=activePatient(),type=ui.task.type,slot=taskSlots(patient,type).find(entry=>entry.id===slotId);
  if(type==='training'&&configuredTrainingDeferred(patient))return toast('今日已有旧版训练记录，新版从明日起执行，不重复补做。');
  if(!slot||!enabledTasks(patient).includes(type))return toast('当前任务不可执行。');
  if(slotDone(patient,type,slotId)||(type==='training'&&legacyTrainingDone(patient)))return toast('该任务已记录，无需重复完成。');
  ui.task.slot=slotId;
  if(type==='training')training();else if(type==='nutrition')nutrition();else survey();
}
function showTaskPage(title,subtitle,body,footer=''){
  ui.patientPage='task';ui.task.body={title,subtitle,body,footer};render();
  document.querySelector('.task-page-heading')?.focus();
}
function finishPatientTask(){ui.task.slot=null;ui.task.body=null;persist();}
function patientTaskPage(patient){
  if(!ui.task)return patientHome(patient);
  const type=ui.task.type,content=ui.task.body,slots=taskSlots(patient,type);
  const records=patientRecords(patient.id).filter(record=>record.date===TODAY&&record.type===type);
  const complete=slots.filter(slot=>slotDone(patient,type,slot.id)).length;
  const back=content?'task-overview':'patient-page';
  return `<section class="task-page"><button class="link task-back" data-action="${back}" data-value="home">← ${content?'返回'+serviceTypes[type]:'返回今日'}</button><div class="phone-page-title"><div class="eyebrow">${displayDate(TODAY)} · 计划 V${patient.plan.version}</div><h2 class="task-page-heading" tabindex="-1">${content?escapeHtml(content.title):serviceTypes[type]}</h2><p>${content?escapeHtml(content.subtitle):{training:'按医护设置的动作与时段，从容完成训练。',nutrition:'计划可按全天或餐次设置，实际摄入仍按餐记录。',survey:'记录今天的感受，让团队更了解你。'}[type]}</p></div>${content?`<div class="task-context">${escapeHtml(taskSlot()?.label||'今日随访')} · ${escapeHtml(patient.name)}</div><div id="task-form-error" class="validation-error" role="alert"></div><div class="task-page-content">${type==='nutrition'?carePatientMeal(ui.task.slot):''}${content.body}</div><footer class="task-page-actions">${content.footer}</footer>`:`<div class="task-summary"><span class="task-icon">${icon(type==='nutrition'?'food':type==='survey'?'chat':'training')}</span><div><h3>${type==='training'?(patient.plan.actions?`每日 ${slots.length} 项 · 共 ${slots.reduce((total,slot)=>total+slot.sets*slot.reps,0)} 次`:`今日共 ${patient.plan.sets} 组 × ${patient.plan.reps} 次`):type==='nutrition'?`三餐已记录 ${complete} / 3`:(taskDone(patient,type)?'今天已反馈':'今天还未反馈')}</h3><p>${type==='training'?'仅展示已安排的训练；全天总量为一天合计，不需每个时段重复做':type==='nutrition'?'记录完成不代表摄入达标':'如有新的不适，随时主动反馈'}</p></div></div>${type==='nutrition'?carePatientMeal('allDay',patient):''}${type==='training'&&configuredTrainingDeferred(patient)?'<div class="notice">今日已有旧版训练记录，以下是新版本安排，从明日起执行。今日请勿重复补做；旧记录继续保留。</div>':''}${type==='training'&&patient.plan.actions?'<p class="task-hint">各项训练由医护选择与配置。示例画面不是动作指导；不适时停止，不为打卡补做。</p>':type==='training'?'<p class="task-hint">原型将每日组数拆分至上午、下午；时段为交互示例，实际安排需医护确认。不适时停止，不为打卡补做。</p>':type==='nutrition'?'<p class="task-hint">计划份量与记录任务分开：全天量只展示一次，实际用餐仍分餐记录，不自动拆分全天量。没吃也可以填写原因；不要为了完成记录改变饮食方案。</p>':'<p class="task-hint">内容直接提供给团队，不自动诊断。紧急情况不要等待线上回复。</p>'}${slots.map(slot=>{const record=slotRecord(patient,type,slot.id),done=slotDone(patient,type,slot.id)||(type==='training'&&legacyTrainingDone(patient))||(type==='survey'&&taskDone(patient,type));return `<article class="task-slot"><div class="between"><h3>${escapeHtml(slot.label)}</h3>${sourceTag(done?'已记录':record?.status||'待记录',done?'green':'orange')}</div><p>${type==='training'?`${slot.sets} 组 × ${slot.reps||patient.plan.reps} 次 · ${slot.time}`:type==='nutrition'?`${slot.time} · 食物、份量与实食比例`:'不适情况 · 执行困难 · 补充说明'}</p>${type==='nutrition'?carePatientMeal(slot.id,patient):type==='training'&&slot.note?`<p>${escapeHtml(slot.note)}</p>`:''}${record?`<p class="slot-detail">${escapeHtml(record.detail)}</p><small>${record.time} · 计划 V${record.version}</small>`:''}${type==='training'&&configuredTrainingDeferred(patient)?'<p class="task-hint">新版训练明日起执行</p>':!done?`<button class="btn block ${type==='nutrition'?'':'primary'}" data-action="task-slot" data-value="${slot.id}">${record?'继续记录':type==='training'?'开始这次训练':type==='nutrition'?'记录'+slot.label:'填写今日随访'} ${icon('arrow')}</button>`:'<p class="slot-complete">✓ 记录已保存，无需重复打卡</p>'}</article>`;}).join('')}${records.some(record=>!record.slot)?`<div class="notice">有 ${records.filter(record=>!record.slot).length} 条旧版记录未区分餐次 / 时段。${type==='training'&&legacyTrainingDone(patient)?'今日已有整日训练完成记录，无需再次训练。':type==='nutrition'?'保留原记录，不推定三餐均已完成。':'原始记录继续保留。'}</div>`:''}<div class="task-page-actions"><button class="btn block" data-action="discomfort">${icon('chat')}有新的不适，主动反馈</button></div>`}</section>`;
}
function feedbackLevel(message){return ['high','medium','normal'].includes(message.level)?message.level:'normal';}
function clinicianFeedback(patient){return state.messages.filter(message=>message.pid===patient.id&&message.source==='医护确认'&&!message.read&&!message.dismissed).sort((left,right)=>({high:0,medium:1,normal:2}[feedbackLevel(left)]-{high:0,medium:1,normal:2}[feedbackLevel(right)]));}
function clinicianFeedbackBanner(patient){
  const messages=clinicianFeedback(patient),message=messages[0];if(!message)return '';
  const level=feedbackLevel(message),label={high:'紧急 · 请优先查看',medium:'重要 · 医护有新反馈',normal:'医护有新反馈'}[level];
  return `<section class="clinician-feedback ${level}" aria-label="医护新反馈"><span class="feedback-icon">${icon(level==='high'?'alert':level==='medium'?'flag':'bell')}</span><div><strong>${label}</strong><p>${escapeHtml(message.title)}</p><button class="link" data-action="view-feedback" data-id="${message.id}">查看医护反馈${messages.length>1?` · 共 ${messages.length} 条`:''} →</button>${level==='high'?'<small>紧急不适请及时就医，不要等待线上回复。</small>':''}</div><button class="icon-btn" aria-label="关闭本次反馈提示" data-action="dismiss-feedback" data-id="${message.id}">${icon('close')}</button></section>`;
}
function patientFeedbackPage(patient){
  const message=state.messages.find(entry=>entry.id===ui.feedbackId&&entry.pid===patient.id);
  if(!message)return patientMessages(patient);
  return `<button class="link task-back" data-action="patient-page" data-value="home">← 返回今日</button><div class="phone-page-title"><h2>医护给你的反馈</h2><p>已查看本次提示 · 消息仍会保留</p></div><article class="message-card">${sourceTag({high:'紧急',medium:'重要',normal:'常规'}[feedbackLevel(message)],feedbackLevel(message)==='high'?'red':'green')}<h3 style="margin-top:15px">${escapeHtml(message.title)}</h3><p>${escapeHtml(message.text)}</p><small>${escapeHtml(message.author)} · ${escapeHtml(message.time)}</small></article><div class="notice">关闭提示或查看消息，不会关闭待跟进事项。紧急情况请及时线下就医，不等待线上回复。</div><button class="btn block" data-action="patient-page" data-value="messages">查看全部消息与跟进</button>`;
}
