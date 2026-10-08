const foodReferenceSeeds = {
  'food-1':{name:'米饭',unit:'克',reference:{label:'小碗',shape:'bowl',amount:150,capacity:250,description:'示例碗口直径约 11 cm、高约 6 cm；熟米饭松装至平口，不压实、不堆尖。',basis:'熟米饭可食部，不含额外油脂或配料',energyBase:100,energy:130}},
  'food-2':{name:'鸡蛋',unit:'个',reference:{label:'个',shape:'piece',amount:1,capacity:null,description:'示例为一枚中等大小鸡蛋，去壳可食部约 50 g；不是一碗鸡蛋。',basis:'去壳熟鸡蛋，不含额外烹调用油',energyBase:1,energy:78}},
  'food-3':{name:'西兰花',unit:'克',reference:{label:'小碗',shape:'bowl',amount:100,capacity:250,description:'示例碗口直径约 11 cm、高约 6 cm；焯熟沥水后切小朵、自然摆放至平口，不压实。',basis:'焯熟沥水的可食部，不含额外油脂或酱料',energyBase:100,energy:35}},
  'food-4':{name:'牛奶',unit:'毫升',reference:{label:'杯',shape:'cup',amount:250,capacity:250,description:'示例直筒量杯，直径约 6 cm、高约 9 cm；以 250 mL 刻度为准，不是任意家用杯。',basis:'示例原味牛奶，具体产品以包装标示为准',energyBase:100,energy:60}}
};
function ensureFoodReferences() {
  for(const item of state.catalog.foods) {
    const seed=foodReferenceSeeds[item.id];
    if(!Object.hasOwn(item,'reference')&&item.version===1&&seed?.name===item.name&&seed.unit===item.unit) {
      item.reference={...seed.reference,source:'原型模拟系数，未经营养数据库校验；上线前由院方营养师核定'};
      item.version+=1;
    }
  }
}
function foodNumber(value) {
  if(!Number.isFinite(value))return '—';
  if(value>0&&value<0.01)return '<0.01';
  return String(Math.round(value*100)/100);
}
function foodEstimate(item,amount) {
  const reference=item.reference, validAmount=Number.isFinite(amount)&&amount>0;
  const portions=validAmount&&reference&&Number.isFinite(reference.amount)&&reference.amount>0?amount/reference.amount:null;
  const kcal=validAmount&&reference&&Number.isFinite(reference.energyBase)&&reference.energyBase>0&&Number.isFinite(reference.energy)&&reference.energy>=0?amount/reference.energyBase*reference.energy:null;
  return {portions,kcal};
}
function foodEquivalent(item,amount) {
  const estimate=foodEstimate(item,amount);
  return estimate.portions===null?'生活份量待配置':`约 ${foodNumber(estimate.portions)} ${escapeHtml(item.reference.label)}`;
}
function foodEnergyText(kcal) {
  return kcal===null?'能量未估算':`约 ${foodNumber(kcal)} kcal`;
}
function foodReferenceForm(item) {
  const reference=item?.reference;
  return `<section class="food-reference-editor"><h3>生活份量与能量参考</h3><label class="checkbox"><input id="food-reference-enabled" type="checkbox" ${reference?'checked':''}>启用参考换算</label><p class="care-caption">按本食材、烹调状态分别维护。容量不能直接当作重量；关闭后显示未配置，不当作 0 kcal。</p><div id="food-reference-fields" ${reference?'':'hidden'}>
    <div class="form-row">${field('生活单位名称',input('food-label',reference?.label||'小碗','text','maxlength="12"'))}${field('容器 / 实物图示',careSelect('food-shape',{bowl:'碗',cup:'杯',piece:'个 / 件'},reference?.shape||'bowl'))}</div>
    ${field('每一生活单位对应的计量数值',input('food-amount',reference?.amount??'','number','min="0.01" max="3000" step="0.01"'),'使用上方计量单位。例如单位为克：1 小碗米饭 = 150 克；不得把 150 克填写成 150 毫升。')}
    ${field('容器标称容量（mL，可选）',input('food-capacity',reference?.capacity??'','number','min="1" max="5000" step="1"'))}
    ${field('尺寸 / 大小与装盛说明',textarea('food-description',reference?.description||'','例如碗口直径、高度、装到几分满、是否压实；按个计量时说明大小。'))}
    ${field('食物状态与能量口径',input('food-basis',reference?.basis||'','text','maxlength="150"'),'说明生重 / 熟重、可食部，是否含油、糖或酱料。')}
    <div class="form-row">${field('能量基准量（同上方计量单位）',input('food-energy-base',reference?.energyBase??'','number','min="0.01" max="3000" step="0.01"'))}${field('该基准量的能量（kcal）',input('food-energy',reference?.energy??'','number','min="0" max="10000" step="0.01"'))}</div>
    <p class="care-caption">两项能量数值可一起留空，表示未估算；明确为 0 才填写 0。</p>
    ${field('参考来源 / 核验说明',textarea('food-source',reference?.source||'','原型请注明模拟数据；正式内容需院方填写数据库、产品标签或核验依据。'))}
    <div class="notice">修改计量单位时，必须同时重新核对份量与能量基准。填写来源不代表系统已完成核验。</div></div></section>`;
}
function readFoodReference() {
  if(!checked('food-reference-enabled'))return null;
  return {label:read('food-label'),shape:read('food-shape'),amount:Number(read('food-amount')),capacity:read('food-capacity')===''?null:Number(read('food-capacity')),description:read('food-description'),basis:read('food-basis'),energyBase:read('food-energy-base')===''?null:Number(read('food-energy-base')),energy:read('food-energy')===''?null:Number(read('food-energy')),source:read('food-source')};
}
function foodReferenceValidation(reference) {
  if(!reference)return '';
  if(!reference.label||reference.label.length>12||!['bowl','cup','piece'].includes(reference.shape))return '请填写 1–12 字生活单位并选择图示。';
  if(!Number.isFinite(reference.amount)||reference.amount<=0||reference.amount>3000)return '每一生活单位的计量数值须大于 0 且不超过 3000。';
  if(reference.capacity!==null&&(!Number.isInteger(reference.capacity)||reference.capacity<1||reference.capacity>5000))return '容器容量须为 1–5000 mL 的整数，或留空。';
  if(reference.description.length<6||reference.description.length>500||!reference.basis||reference.basis.length>150||reference.source.length<6||reference.source.length>500)return '请补充大小说明和来源（各 6–500 字），以及食物状态口径（1–150 字）。';
  if(reference.energyBase===null&&reference.energy===null)return '';
  if(!Number.isFinite(reference.energyBase)||reference.energyBase<=0||reference.energyBase>3000||!Number.isFinite(reference.energy)||reference.energy<0||reference.energy>10000)return '能量基准与 kcal 须成对填写：基准量大于 0、不超过 3000；能量为 0–10000。以上仅为原型输入范围。';
  return '';
}
function foodVessel(shape,half=false) {
  const outline=shape==='cup'?'M27 21H83L77 77H33Z':shape==='piece'?'M55 18C42 18 27 48 29 62C31 87 79 87 81 62C83 48 68 18 55 18Z':'M13 32H97C91 66 77 78 55 78S19 66 13 32Z';
  return `<svg class="food-vessel" viewBox="0 0 110 92" aria-hidden="true"><path d="${outline}" fill="${half?'#edf3e5':'#c8ddae'}" stroke="#71896a" stroke-width="2"/>${half?'<path d="M38 59h34" stroke="#8aa975" stroke-width="12" stroke-linecap="round"/>':''}${shape==='bowl'?'<ellipse cx="55" cy="32" rx="42" ry="9" fill="#f5f8ef" stroke="#71896a" stroke-width="2"/>':''}</svg>`;
}
function foodSizeGuide(item,collapsible=true) {
  const reference=item.reference;if(!reference)return '';
  const container=collapsible?'details':'div';
  return `<${container} class="food-size-guide">${collapsible?'<summary>份量大小对照与参考口径</summary>':''}<div class="food-size-pair">${[0.5,1].map(scale=>`<div>${foodVessel(reference.shape,scale===0.5)}<strong>${scale} ${escapeHtml(reference.label)}</strong><span>约 ${foodNumber(reference.amount*scale)} ${escapeHtml(item.unit)}</span></div>`).join('')}</div><p>${reference.capacity?`示例容器：${reference.capacity} mL。`:''}${escapeHtml(reference.description)}</p><p>图示仅表示份数，不按实物尺寸或容积比例绘制。以说明中的同款容器和装盛方式为准，家用碗杯可能不同。</p><p><b>食物口径：</b>${escapeHtml(reference.basis)}</p><p><b>能量基准：</b>${reference.energy===null||reference.energyBase===null?'未配置':`${reference.energyBase} ${escapeHtml(item.unit)} ≈ ${foodNumber(reference.energy)} kcal`}</p><p><b>参考来源：</b>${escapeHtml(reference.source)}</p></${container}>`;
}
function foodReferenceCard(item,amount,detail=true) {
  const estimate=foodEstimate(item,amount);
  return `<div class="food-reference"><div class="food-reference-values"><strong>${foodNumber(amount)} ${escapeHtml(item.unit)} · ${foodEquivalent(item,amount)}</strong><span>${foodEnergyText(estimate.kcal)}</span></div>${item.reference?`<p>${escapeHtml(item.reference.basis)} · 仅为参考，不是推荐摄入量</p><small>${escapeHtml(item.reference.source)}</small>${detail?foodSizeGuide(item):''}`:'<p>当前库 / 计划快照尚无换算依据，不猜测碗数或能量。请由医护重新选用已配置内容。</p>'}</div>`;
}
function foodRowReference(row,compact=false) {
  if(compact)return `<div class="food-row-reference compact" data-food-row="${escapeHtml(row.rowId)}"><div class="food-reference-values"><strong>${foodNumber(row.amount)} ${escapeHtml(row.snapshot.unit)} · ${foodEquivalent(row.snapshot,row.amount)}</strong><span>${foodEnergyText(foodEstimate(row.snapshot,row.amount).kcal)}</span></div><details class="care-row-note"><summary>份量大小对照 / 快捷换算（模拟参考）</summary>${row.snapshot.reference?`${foodSizeGuide(row.snapshot,false)}<p>仅为参考，非推荐摄入量</p><div class="food-portion-buttons">${[0.5,1,1.5,2].map(scale=>careButton('food-portion',`${scale} ${escapeHtml(row.snapshot.reference.label)}`,`${row.rowId}:${scale}`,'small')).join('')}</div>`:'<p>当前快照没有换算依据，请医护核对，不自动推算。</p>'}</details></div>`;
  return `<div class="food-row-reference" data-food-row="${escapeHtml(row.rowId)}">${foodReferenceCard(row.snapshot,row.amount)}${row.snapshot.reference?`<div class="food-portion-buttons"><span>快捷份量</span>${[0.5,1,1.5,2].map(scale=>careButton('food-portion',`${scale} ${escapeHtml(row.snapshot.reference.label)}`,`${row.rowId}:${scale}`,'small')).join('')}</div>`:''}</div>`;
}
function foodEnergyTotal(rows,label='已配置食材合计') {
  if(!rows.length)return `<p class="food-energy-total">${label}：未配置，不估算能量</p>`;
  const values=rows.map(row=>foodEstimate(row.snapshot,row.amount).kcal), known=values.filter(value=>value!==null),missing=values.length-known.length;
  return `<p class="food-energy-total"><strong>${label}：${known.length?foodEnergyText(known.reduce((sum,value)=>sum+value,0)):'能量未估算'}</strong>${missing?`<span>还有 ${missing} 项未估算，不能作为完整合计。</span>`:''}<span>仅汇总所列食材，不代表完整膳食、实际摄入或推荐目标。</span></p>`;
}
function refreshFoodRowReferences() {
  for(const row of careUi.draft?.foods||[]) {
    const container=Array.from(document.querySelectorAll('[data-food-row]')).find(element=>element.dataset.foodRow===row.rowId);
    if(!container)continue;
    const estimate=foodEstimate(row.snapshot,row.amount),values=container.querySelector('.food-reference-values');
    values.innerHTML=`<strong>${foodNumber(row.amount)} ${escapeHtml(row.snapshot.unit)} · ${foodEquivalent(row.snapshot,row.amount)}</strong><span>${foodEnergyText(estimate.kcal)}</span>`;
  }
}
document.addEventListener('change',event=>{
  if(event.target.id==='food-reference-enabled')document.getElementById('food-reference-fields').hidden=!event.target.checked;
  if(event.target.id==='catalog-unit'&&document.getElementById('food-reference-fields')) {
    for(const id of ['food-amount','food-energy-base','food-energy'])document.getElementById(id).value='';
    toast('计量单位已变更，请重新填写份量和能量基准，并核对说明。');
  }
});
