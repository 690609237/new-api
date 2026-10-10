const shopCategories=['营养餐','减脂餐','医疗器械 / 穿戴设备','医疗保健器械','训练课程'];
const shopStatuses=['待支付','待发货','已发货','服务已开通','已完成','已取消','已退款'];
let shopUi={category:'全部',search:'',orderFilter:'全部',orderSearch:'',productId:null,orderId:null};
const shopMoney=cents=>`¥${(cents/100).toFixed(2)}`;
const shopButton=(action,label,id='',style='')=>`<button class="btn ${style}" data-action="shop-${action}" data-id="${escapeHtml(id)}">${label}</button>`;
const shopPatient=()=>ui.role==='patient';
const shopAdmin=()=>careAdmin();
function ensureShopData(){
  if(state.commerce)return;
  state.commerce={products:[
    {id:'goods-meal',name:'居家营养餐 · 单次配送',category:'营养餐',kind:'physical',price:3800,stock:80,description:'示例套餐，含主食、蛋白类与蔬菜。购买前与服务团队核对个体饮食限制。',spec:'1 人份 / 1 餐',usage:'配料与过敏原以交付说明为准；非所有术后患者均适用。',policy:'餐食未发货可申请退款；配送后的质量问题可申请售后，由人工核实。',active:true},
    {id:'goods-light',name:'轻食搭配餐 · 单次配送',category:'减脂餐',kind:'physical',price:3200,stock:60,description:'用于展示院方餐食服务，不承诺减脂效果，不替代营养方案。',spec:'1 人份 / 1 餐',usage:'术后恢复期间不默认适合控制摄入，请先咨询团队。',policy:'餐食未发货可申请退款；配送后的问题由人工核实处理。',active:true},
    {id:'goods-band',name:'居家活动记录手环',category:'医疗器械 / 穿戴设备',kind:'physical',price:29900,stock:20,description:'虚构穿戴设备，演示活动记录用途；不接入真实监测，不用于诊断。',spec:'标准款 / 1 台',usage:'是否属于医疗器械、适用范围与资质待院方审核；示例不是获批产品。',policy:'支持提交退货退款申请，使用状况、质量问题由人工核对。',active:true},
    {id:'goods-aid',name:'居家辅助训练用品',category:'医疗保健器械',kind:'physical',price:6900,stock:30,description:'虚构辅助用品，不替代康复治疗；由医护核对是否适用。',spec:'基础套装 / 1 套',usage:'使用方式、禁忌与清洁要求需核对产品说明和医护安排。',policy:'支持提交退货退款申请，由人工核对商品状态。',active:true},
    {id:'goods-course',name:'居家训练入门课程',category:'训练课程',kind:'digital',price:9900,stock:100,description:'4 节示例课程，仅演示购买和服务开通，无真实教学视频。',spec:'4 节 / 开通后 30 天',usage:'购买不自动加入训练计划；实际执行须医护确认。',policy:'未开始或已开始均可提出退款申请，按实际服务情况人工审核。',active:true}
  ],orders:[],gateway:{enabled:true,wechat:true,alipay:true,outcome:'success',refundOutcome:'success',merchant:'院方居家服务 · 演示商户'}};
  const examples=[
    {id:'demo-order-meal',productId:'goods-meal',status:'待发货',pid:state.patients[0].id},
    {id:'demo-order-band',productId:'goods-band',status:'已发货',pid:state.patients[0].id},
    {id:'demo-order-course',productId:'goods-course',status:'服务已开通',pid:(state.patients[1]||state.patients[0]).id}
  ];
  for(const example of examples){
    const product=state.commerce.products.find(item=>item.id===example.productId);
    state.commerce.orders.push({id:example.id,pid:example.pid,product:careClone(product),quantity:1,total:product.price,address:product.kind==='digital'?'线上服务，无需配送':'演示收件人 / 示例联系电话 / 示例市示例路 1 号',status:example.status,created:TODAY,history:[{time:`${TODAY} 09:00`,text:'预置演示订单，模拟支付成功，非真实交易'},...(example.status==='已发货'?[{time:`${TODAY} 10:00`,text:'预置演示：已发货，患者申请核实商品问题，等待售后审核'}]:[])],shipping:example.status==='已发货'?{carrier:'演示快递',tracking:'DEMO000001'}:null,afterSale:example.status==='已发货'?{status:'待审核',reason:'虚构设备出现问题，申请退货并核实',reply:'',appeal:'',returnShipping:null}:null,payment:{id:'mock-pay-'+example.id,channel:'微信支付'},stockReturned:false});
    product.stock-=1;
  }
}
function shopNotice(){return '<div class="notice">自愿购买 · 不购买不影响原有照护服务。商品不是处方，购买不自动改变医护计划。所有价格、支付、物流与退款均为本地模拟，请勿填写真实身份、地址或支付凭据。</div>';}
function shopLibraryTabs(){return `<div class="care-tabs">${careButton('library-tab','动作库','actions')}${careButton('library-tab','食材库','foods')}${careButton('library-tab','商品库','products','primary')}</div>`;}
function shopFilters(){return `<div class="shop-filters">${['全部',...shopCategories].map(category=>shopButton('category',escapeHtml(category),category,shopUi.category===category?'primary':'')).join('')}</div><label class="shop-search">搜索商品<input id="shop-search" value="${escapeHtml(shopUi.search)}" placeholder="名称 / 商品说明，回车搜索"></label>`;}
function shopProducts(){return state.commerce.products.filter(product=>(!shopPatient()||product.active)&&(shopUi.category==='全部'||product.category===shopUi.category)&&`${product.name} ${product.description}`.includes(shopUi.search));}
function shopProductCards(){return `<div class="shop-grid">${shopProducts().map(product=>`<article class="shop-card"><div class="shop-art ${product.kind==='digital'?'course':''}">${icon(product.kind==='digital'?'training':product.category.includes('餐')?'food':'heart')}<span>${escapeHtml(product.category)}</span></div><div class="shop-card-body"><div class="between">${sourceTag(product.kind==='digital'?'线上服务':'实物配送','green')}${!shopPatient()?sourceTag(product.active?'已上架':'已下架',product.active?'green':''):''}</div><h3>${escapeHtml(product.name)}</h3><p>${escapeHtml(product.spec)}</p><p>${escapeHtml(product.description)}</p><div class="between"><strong class="shop-price">${shopMoney(product.price)}</strong><small>余量 ${product.stock}</small></div><div class="shop-actions">${shopButton('product','查看详情',product.id)}${shopAdmin()?shopButton('edit-product','编辑 / 上下架',product.id):''}</div></div></article>`).join('')||'<div class="empty">暂无匹配商品，试试其他分类。</div>'}</div>`;}
function renderShopLibrary(){
  ensureShopData();
  return `<div class="page-heading between"><div><div class="eyebrow">CARE SERVICE STORE</div><h1>照护内容库 · 商品库</h1><p>院方服务相关商品，与临床计划独立管理。</p></div>${shopAdmin()?shopButton('edit-product','＋ 新增商品','','primary'):sourceTag('普通医护 · 只读查看')}</div>${shopLibraryTabs()}${shopNotice()}${shopAdmin()?`<div class="shop-actions">${shopButton('admin-orders','订单与售后')}${shopButton('gateway','模拟支付设置')}</div>`:''}${shopFilters()}${shopProductCards()}`;
}
function patientShop(){ensureShopData();return `<div class="phone-page-title between"><div><div class="eyebrow">CARE AT HOME</div><h2>居家服务商城</h2></div>${shopButton('orders','我的订单')}</div>${shopNotice()}${shopFilters()}${shopProductCards()}`;}
function shopMineLinks(){return `<button class="task-card" data-action="shop-store"><span class="task-icon food">${icon('food')}</span><div><h3>居家服务商城</h3><p>餐食 · 器械 · 课程，自愿选购</p></div><span class="arrow">${icon('arrow')}</span></button><button class="task-card" data-action="shop-orders"><span class="task-icon">${icon('list')}</span><div><h3>我的订单与售后</h3><p>支付 · 配送 · 退货与申诉</p></div><span class="arrow">${icon('arrow')}</span></button>`;}
function shopProductPage(){
  const product=state.commerce.products.find(item=>item.id===shopUi.productId&&item.active);
  if(!product)return `${shopButton('store','← 返回商城')}<div class="empty">商品已下架，请查看其他商品。</div>`;
  return `${shopButton('store','← 返回商城')}<div class="shop-detail"><div class="shop-art">${icon(product.kind==='digital'?'training':'food')}<span>${escapeHtml(product.category)}</span></div><h2>${escapeHtml(product.name)}</h2><p>${escapeHtml(product.spec)}</p><strong class="shop-price">${shopMoney(product.price)}</strong><p>${escapeHtml(product.description)}</p><h3>适用与使用说明</h3><p>${escapeHtml(product.usage)}</p><h3>售后说明</h3><p>${escapeHtml(product.policy)}</p>${shopNotice()}${product.stock>0?shopButton('checkout','自愿选购 · 确认订单',product.id,'primary block'):'<p>暂时售罄</p>'}</div>`;
}
function shopProductForm(id){
  if(!shopAdmin())return toast('仅管理员可维护商品。');
  const product=state.commerce.products.find(item=>item.id===id)||{name:'',category:shopCategories[0],kind:'physical',price:0,stock:0,description:'',spec:'',usage:'',policy:'',active:false};
  modalContext={kind:'shop-product',id};
  showModal(id?'编辑商品':'新增商品','仅模拟上架；真实商品资质与履约能力须院方核验',`${field('商品名称',input('shop-name',product.name,'text','maxlength="60"'))}<div class="form-row">${field('分类',careSelect('shop-category',Object.fromEntries(shopCategories.map(category=>[category,category])),product.category))}${field('交付方式',careSelect('shop-kind',{physical:'实物配送',digital:'线上课程 / 服务'},product.kind))}</div>${field('规格 / 服务周期',input('shop-spec',product.spec,'text','maxlength="100"'))}<div class="form-row">${field('单价（元）',input('shop-price',(product.price/100).toFixed(2),'number','min="0.01" max="100000" step="0.01"'))}${field('可售库存 / 名额',input('shop-stock',product.stock,'number','min="0" max="100000"'))}</div>${field('商品说明',textarea('shop-description',product.description))}${field('适用限制 / 资质核对说明',textarea('shop-usage',product.usage))}${field('售后政策说明',textarea('shop-policy',product.policy))}<label class="checkbox"><input id="shop-active" type="checkbox" ${product.active?'checked':''}><span>上架展示（原型模拟，非真实资质审批）</span></label>`,shopButton('save-product','保存商品','','primary'));
}
function shopSaveProduct(){
  if(!shopAdmin()||modalContext.kind!=='shop-product')return;
  const price=Number(read('shop-price')),stock=Number(read('shop-stock'));
  const values={name:read('shop-name'),category:read('shop-category'),kind:read('shop-kind'),spec:read('shop-spec'),price:Math.round(price*100),stock,description:read('shop-description'),usage:read('shop-usage'),policy:read('shop-policy'),active:checked('shop-active')};
  if(!values.name||values.name.length>60||!values.spec||values.spec.length>100||[values.description,values.usage,values.policy].some(text=>text.length<6||text.length>1500))return formError('请填写名称、规格和至少 6 字的商品、适用及售后说明（说明最多 1500 字）。');
  if(!shopCategories.includes(values.category)||!['physical','digital'].includes(values.kind)||!Number.isFinite(price)||price<=0||price>100000||Math.abs(price*100-values.price)>0.00001||!read('shop-stock')||!Number.isInteger(stock)||stock<0||stock>100000)return formError('请核对分类、交付方式、单价（最多两位小数）和整数库存。');
  const existing=state.commerce.products.find(product=>product.id===modalContext.id);
  if(existing)Object.assign(existing,values);else state.commerce.products.push({id:uuid('goods'),...values});
  modalContext={};closeModal();persist();toast('商品已保存，历史订单保持下单时的快照');
}
function shopCheckout(id){
  if(!shopPatient())return;
  const product=state.commerce.products.find(item=>item.id===id&&item.active&&item.stock>0);if(!product)return toast('商品已下架或售罄。');
  modalContext={kind:'shop-checkout',id};
  showModal('确认订单','演示下单，不产生真实扣款',`<h3>${escapeHtml(product.name)}</h3><p>${shopMoney(product.price)} / ${escapeHtml(product.spec)} · 运费 ¥0.00（模拟）</p>${field('数量',input('shop-quantity',1,'number',product.kind==='digital'?'min="1" max="1"':'min="1" max="10"'))}${product.kind==='physical'?field('收货信息（仅虚构内容）',textarea('shop-address','演示收件人 / 演示联系电话 / 示例市示例路 1 号')):'<div class="notice">付款成功后为当前演示患者开通课程，不需要物流；购买不会修改训练计划。</div>'}<p>${escapeHtml(product.policy)}</p><label class="checkbox"><input id="shop-consent" type="checkbox"><span>我已查看商品及售后说明，自愿购买；不购买不影响照护服务。</span></label>`,shopButton('place-order','创建待支付订单','','primary'));
}
function shopAudit(order,text){order.history.push({time:`${TODAY} ${timeLabel()}`,text});}
function shopOwnOrder(id){const order=state.commerce.orders.find(item=>item.id===id);return order&&(shopAdmin()||(shopPatient()&&order.pid===activePatient().id))?order:null;}
function shopPlaceOrder(){
  if(!shopPatient()||modalContext.kind!=='shop-checkout')return;
  const product=state.commerce.products.find(item=>item.id===modalContext.id),quantity=Number(read('shop-quantity')),address=read('shop-address');
  if(!product?.active||!Number.isInteger(quantity)||quantity<1||quantity>10||quantity>product.stock)return formError('请核对商品状态及数量（1–10，且不超过库存）。');
  if(product.kind==='digital'&&quantity!==1)return formError('线上课程每笔订单仅开通当前患者 1 份服务。');
  if(product.kind==='physical'&&(address.length<6||address.length>300))return formError('请填写 6–300 字的虚构收货信息。');
  if(!checked('shop-consent'))return formError('请先确认商品说明与自愿购买。');
  const order={id:uuid('order'),pid:activePatient().id,product:careClone(product),quantity,total:product.price*quantity,address:product.kind==='physical'?address:'线上服务，无需配送',status:'待支付',created:TODAY,history:[],shipping:null,afterSale:null,payment:null,stockReturned:false};
  product.stock-=quantity;shopAudit(order,'患者创建订单，锁定价格与库存，尚未支付');state.commerce.orders.unshift(order);
  shopUi.orderId=order.id;ui.patientPage='shopOrder';modalContext={};closeModal();persist();
}
function shopOrderFilters(){return `<div class="shop-filters">${['全部',...shopStatuses,'售后处理中'].map(status=>shopButton('order-filter',status,status,shopUi.orderFilter===status?'primary':'')).join('')}</div><label class="shop-search">搜索订单<input id="shop-order-search" value="${escapeHtml(shopUi.orderSearch)}" placeholder="订单号 / 商品名称"></label>`;}
function shopAfterOpen(order){return order.afterSale&&!['已退款','已驳回'].includes(order.afterSale.status);}
function shopOrders(admin=false){
  ensureShopData();if(admin&&!shopAdmin())return '<div class="empty">仅管理员可查看订单与售后。</div>';if(!admin&&!shopPatient())return '';
  const orders=state.commerce.orders.filter(order=>(admin||order.pid===activePatient().id)&&(shopUi.orderFilter==='全部'||(shopUi.orderFilter==='售后处理中'?shopAfterOpen(order):order.status===shopUi.orderFilter))&&`${order.id} ${order.product.name}`.includes(shopUi.orderSearch));
  return `<div class="${admin?'page-heading':'phone-page-title'} between"><div><div class="eyebrow">ORDERS & AFTERCARE</div><h${admin?'1':'2'}>${admin?'订单与售后':'我的订单'}</h${admin?'1':'2'}></div>${admin?shopButton('gateway','模拟支付设置'):shopButton('store','逛商城')}</div>${shopNotice()}${shopOrderFilters()}<div class="shop-order-list">${orders.map(order=>`<article class="shop-order-card"><div class="between">${sourceTag(order.status,'green')}<small>${escapeHtml(order.created)}</small></div><h3>${escapeHtml(order.product.name)}</h3><p>${escapeHtml(order.product.spec)} × ${order.quantity} · ${shopMoney(order.total)}</p><small>${escapeHtml(order.id)}${admin?' · '+escapeHtml(patientById(order.pid)?.name||'演示患者'):''}</small>${order.afterSale?`<p>${sourceTag('售后 · '+order.afterSale.status,'orange')}</p>`:''}<div class="shop-actions">${shopButton('order','查看 / 处理',order.id)}</div></article>`).join('')||'<div class="empty">暂无匹配订单。可在患者商城创建一笔模拟订单体验流程。</div>'}</div>`;
}
function renderShopOrders(){return shopOrders(true);}
function patientShopOrders(){return shopOrders(false);}
function shopOrderActions(order){
  if(shopAdmin())return `${order.status==='待发货'&&!shopAfterOpen(order)?shopButton('ship','填写快递 / 发货',order.id,'primary'):''}${order.status==='已发货'&&!shopAfterOpen(order)?shopButton('ship','更正物流信息',order.id):''}${order.afterSale?['待审核','申诉待复核'].includes(order.afterSale.status)?shopButton('review','审核售后 / 申诉',order.id,'primary'):order.afterSale.status==='待收退货'?shopButton('return-received','确认收到退货',order.id):order.afterSale.status==='待退款'?shopButton('refund','模拟原路退款',order.id,'primary'):'':''}`;
  return `${order.status==='待支付'?shopButton('pay','模拟付款',order.id,'primary')+shopButton('cancel','取消订单',order.id):''}${['待支付','待发货'].includes(order.status)&&order.product.kind==='physical'&&!shopAfterOpen(order)?shopButton('address','修改收货信息',order.id):''}${['已发货','服务已开通'].includes(order.status)&&!shopAfterOpen(order)?shopButton('complete',order.product.kind==='digital'?'确认服务完成':'确认收货',order.id):''}${['待发货','已发货','服务已开通','已完成'].includes(order.status)&&!order.afterSale?shopButton('after-sale','申请退款 / 退货',order.id):''}${order.afterSale?.status==='待寄回'?shopButton('return','填写退货物流',order.id,'primary'):''}${order.afterSale?.status==='已驳回'?shopButton('appeal','补充说明并申诉',order.id):''}`;
}
function shopOrderPage(){
  const order=shopOwnOrder(shopUi.orderId);if(!order)return '<div class="empty">订单不存在或当前身份无权查看。</div>';
  const after=order.afterSale;
  return `${shopButton(shopAdmin()?'admin-orders':'orders','← 返回订单列表')}<article class="shop-detail"><div class="between"><h2>订单详情</h2>${sourceTag(order.status,'green')}</div><small>${escapeHtml(order.id)}</small><h3>${escapeHtml(order.product.name)}</h3><p>${escapeHtml(order.product.spec)} × ${order.quantity} · 订单金额 ${shopMoney(order.total)}</p><p>${order.status==='已退款'?'已模拟全额退款':order.payment?'已模拟支付':'尚未支付'}${order.refundId?' · '+escapeHtml(order.refundId):''}</p><p>支付：${order.payment?escapeHtml(order.payment.channel)+' · 模拟成功 · '+escapeHtml(order.payment.id):'尚未支付'} · 运费 ¥0.00（模拟）</p><h3>${order.product.kind==='digital'?'服务交付':'收货信息'}</h3><p>${escapeHtml(order.address)}</p>${order.product.kind==='digital'?`<p>${order.status==='已退款'?'已退款，模拟服务权益已关闭。':order.status==='已取消'?'订单已取消，未开通服务。':order.payment?'服务已模拟开通，无真实视频；服务内容与期限以本订单规格为准。':'付款后模拟开通课程。'}</p>`:''}${order.shipping?`<div class="notice">配送：${escapeHtml(order.shipping.carrier)} · ${escapeHtml(order.shipping.tracking)}<br>仅维护运单信息，无真实物流轨迹。</div>`:''}<h3>下单时的售后说明</h3><p>${escapeHtml(order.product.policy)}</p>${after?`<section class="shop-after"><h3>售后 / 申诉 · ${escapeHtml(after.status)}</h3><p>申请原因：${escapeHtml(after.reason)}</p>${after.appeal?`<p>申诉补充：${escapeHtml(after.appeal)}</p>`:''}${after.reply?`<p>管理员回复：${escapeHtml(after.reply)}</p>`:''}${['待寄回','待收退货'].includes(after.status)?'<p>退货地址：示例市售后中心 1 号（虚构，请勿实际寄件）</p>':''}${after.returnShipping?`<p>退货物流：${escapeHtml(after.returnShipping.carrier)} · ${escapeHtml(after.returnShipping.tracking)}</p>`:''}<p>申请金额 ${shopMoney(order.total)} · 原型仅支持整单售后</p></section>`:''}<div class="shop-actions">${shopOrderActions(order)}</div><h3>订单处理记录</h3><div class="audit-list">${order.history.slice().reverse().map(entry=>`<div><small>${escapeHtml(entry.time)}</small><p>${escapeHtml(entry.text)}</p></div>`).join('')}</div><p class="task-hint">售后未达成一致可从驳回结果继续申诉，由管理员复核。商品使用引起的不适请另行联系医护；订单售后不替代临床跟进。</p></article>`;
}
function shopGateway(){
  if(!shopAdmin())return '<div class="empty">仅管理员可设置支付网关。</div>';
  const gateway=state.commerce.gateway;
  return `<div class="page-heading"><div class="eyebrow">PAYMENT SANDBOX</div><h1>支付网关 · 模拟设置</h1><p>仅供验证成功、失败与退款流程，无真实支付连接。</p></div><section class="shop-detail"><div id="form-error" class="validation-error" role="alert"></div><div class="notice">不接收 API 密钥、证书或真实商户号。真实支付签名、服务端回调验签与资金对账必须在后续服务端实现，前端状态不能作为收款依据。</div>${field('演示商户名称',input('shop-merchant',gateway.merchant,'text','maxlength="60"'))}<label class="checkbox"><input id="shop-enabled" type="checkbox" ${gateway.enabled?'checked':''}>启用模拟支付入口</label><label class="checkbox"><input id="shop-wechat" type="checkbox" ${gateway.wechat?'checked':''}>微信支付（模拟）</label><label class="checkbox"><input id="shop-alipay" type="checkbox" ${gateway.alipay?'checked':''}>支付宝（模拟）</label>${field('支付模拟结果',careSelect('shop-outcome',{success:'成功',failure:'失败，可重试'},gateway.outcome))}${field('退款模拟结果',careSelect('shop-refund-outcome',{success:'成功',failure:'失败，保留待退款'},gateway.refundOutcome))}${shopButton('save-gateway','保存模拟设置','','primary')}</section>`;
}
function shopOperationForm(action,id){
  const order=shopOwnOrder(id);if(!order)return toast('当前身份无权访问该订单。');
  const allowed=shopAdmin()?{ship:['待发货','已发货'].includes(order.status)&&!shopAfterOpen(order),review:['待审核','申诉待复核'].includes(order.afterSale?.status),refund:order.afterSale?.status==='待退款','return-received':order.afterSale?.status==='待收退货'}:{pay:order.status==='待支付',cancel:order.status==='待支付',address:['待支付','待发货'].includes(order.status)&&order.product.kind==='physical'&&!shopAfterOpen(order),complete:['已发货','服务已开通'].includes(order.status)&&!shopAfterOpen(order),'after-sale':['待发货','已发货','服务已开通','已完成'].includes(order.status)&&!order.afterSale,return:order.afterSale?.status==='待寄回',appeal:order.afterSale?.status==='已驳回'};
  if(!allowed[action])return toast('订单状态已变化或无权操作，请重新查看。');
  modalContext={kind:'shop-operation',action,id};
  let body='',title='确认操作';
  if(action==='pay'){
    const gateway=state.commerce.gateway;
    if(!gateway.enabled||(!gateway.wechat&&!gateway.alipay))return toast('模拟支付暂未开放，可保留订单稍后支付。');
    title='模拟付款';body=`<p>演示商户：${escapeHtml(gateway.merchant)}</p><p>应付 ${shopMoney(order.total)}，不会扣款。</p>${field('支付渠道',careSelect('shop-channel',{...(gateway.wechat?{wechat:'微信支付（模拟）'}:{}),...(gateway.alipay?{alipay:'支付宝（模拟）'}:{})},gateway.wechat?'wechat':'alipay'))}`;
  }
  if(['ship','return'].includes(action)){title=action==='ship'?'填写配送物流':'填写退货物流';body=`${field('快递公司',input('shop-carrier',action==='ship'?order.shipping?.carrier||'演示快递':'演示快递','text','maxlength="40"'))}${field('快递单号',input('shop-tracking',action==='ship'?order.shipping?.tracking||'DEMO20261009001':'RETURN20261009001','text','maxlength="50"'))}`;}
  if(action==='address'){title='修改虚构收货信息';body=field('收货信息',textarea('shop-address',order.address));}
  if(['after-sale','appeal','review'].includes(action)){
    title={ 'after-sale':'申请整单退款 / 退货',appeal:'提交售后申诉',review:'人工审核售后 / 申诉'}[action];
    body=`<p>${shopMoney(order.total)} · ${escapeHtml(order.product.policy)}</p>${action==='review'?field('处理决定',careSelect('shop-decision',{approve:'同意申请',reject:'驳回并说明原因'},'approve')):''}${field(action==='review'?'处理依据与回复':'问题说明 / 补充依据',textarea('shop-reason',''))}<p class="task-hint">至少填写 6 字。仅输入虚构说明，不上传真实健康资料。实物已发货时同意申请会进入退货流程；未发货或线上服务进入待退款。</p>`;
  }
  if(action==='cancel')body='<p>取消未支付订单并释放预留库存，不影响照护服务。</p>';
  if(action==='complete')body=`<p>请确认${order.product.kind==='digital'?'服务已完成':'已收到商品'}，确认后仍可申请售后。</p>`;
  if(action==='return-received')body='<p>请核对退回的商品。确认收到后进入待退款，不会立即退款。</p>';
  if(action==='refund')body=`<p>向原模拟支付渠道退款 ${shopMoney(order.total)}；失败时保留待退款，允许重试。</p>`;
  showModal(title,'所有操作只影响本地模拟订单',body,shopButton('confirm-operation','确认','','primary'));
}
function shopRestoreStock(order){
  if(order.stockReturned)return;
  const product=state.commerce.products.find(item=>item.id===order.product.id);if(product)product.stock+=order.quantity;
  order.stockReturned=true;
}
function shopConfirmOperation(){
  if(modalContext.kind!=='shop-operation')return;
  const {action,id}=modalContext,order=shopOwnOrder(id);if(!order)return toast('无权操作该订单。');
  const admin=shopAdmin(),patient=shopPatient(),after=order.afterSale;
  const reason=read('shop-reason');
  if(['after-sale','appeal','review'].includes(action)&&(reason.length<6||reason.length>1500))return formError('请填写 6–1500 字的具体说明。');
  let result='';
  if(patient&&action==='pay'&&order.status==='待支付'){
    const gateway=state.commerce.gateway,channel=read('shop-channel');
    if(!gateway.enabled||!['wechat','alipay'].includes(channel)||!gateway[channel])return formError('该支付渠道当前不可用。');
    if(gateway.outcome==='failure'){shopAudit(order,'模拟支付失败，订单仍待支付，可重试');closeModal();persist();return toast('模拟支付失败，未扣款，可重试');}
    order.payment={id:uuid('mock-pay'),channel:channel==='wechat'?'微信支付':'支付宝'};
    order.status=order.product.kind==='digital'?'服务已开通':'待发货';result='模拟支付成功，'+order.status;
  }else if(patient&&action==='cancel'&&order.status==='待支付'){order.status='已取消';shopRestoreStock(order);result='患者取消未支付订单，已释放库存';
  }else if(patient&&action==='address'&&['待支付','待发货'].includes(order.status)&&order.product.kind==='physical'&&!shopAfterOpen(order)){
    const address=read('shop-address');if(address.length<6||address.length>300)return formError('请填写 6–300 字的虚构收货信息。');order.address=address;result='患者更新了收货信息';
  }else if((admin&&action==='ship'&&['待发货','已发货'].includes(order.status)&&order.product.kind==='physical'&&!shopAfterOpen(order))||(patient&&action==='return'&&after?.status==='待寄回')){
    const carrier=read('shop-carrier'),tracking=read('shop-tracking');if(carrier.length<2||carrier.length>40||!/^[A-Za-z0-9-]{6,50}$/.test(tracking))return formError('请填写快递公司与 6–50 位字母、数字或短横线组成的模拟运单号。');
    if(action==='ship'){order.shipping={carrier,tracking};order.status='已发货';result='管理员维护配送运单：'+carrier+' / '+tracking;}else{after.returnShipping={carrier,tracking};after.status='待收退货';result='患者提交退货物流：'+carrier+' / '+tracking;}
  }else if(patient&&action==='complete'&&['已发货','服务已开通'].includes(order.status)&&!shopAfterOpen(order)){order.status='已完成';result='患者确认收货 / 服务完成';
  }else if(patient&&action==='after-sale'&&['待发货','已发货','服务已开通','已完成'].includes(order.status)&&!after){order.afterSale={status:'待审核',reason,reply:'',appeal:'',returnShipping:null};result='患者申请售后：'+reason;
  }else if(patient&&action==='appeal'&&after?.status==='已驳回'){after.appeal=reason;after.status='申诉待复核';result='患者提交申诉：'+reason;
  }else if(admin&&action==='review'&&['待审核','申诉待复核'].includes(after?.status)){
    const decision=read('shop-decision');if(!['approve','reject'].includes(decision))return formError('请选择处理决定。');
    after.reply=reason;after.status=decision==='reject'?'已驳回':order.product.kind==='physical'&&order.shipping?'待寄回':'待退款';result='管理员审核：'+after.status+'；'+reason;
  }else if(admin&&action==='return-received'&&after?.status==='待收退货'){after.status='待退款';result='管理员确认退货已收到，等待退款；退货库存需管理员检查后手动调整';
  }else if(admin&&action==='refund'&&after?.status==='待退款'){
    if(state.commerce.gateway.refundOutcome==='failure'){shopAudit(order,'模拟退款失败，保留待退款，需重试');closeModal();persist();return toast('模拟退款失败，仍待退款');}
    after.status='已退款';order.status='已退款';order.refundId=uuid('mock-refund');
    if(!order.shipping)shopRestoreStock(order);result='模拟原路退款成功：'+shopMoney(order.total)+' / '+order.refundId;
  }else return toast('状态已变化或当前身份无权操作，请重新打开订单。');
  shopAudit(order,result);closeModal();persist();toast(result);
}
document.addEventListener('click',event=>{
  const button=event.target.closest('[data-action^="shop-"]');if(!button||button.disabled)return;
  event.stopImmediatePropagation();ensureShopData();const action=button.dataset.action.slice(5),id=button.dataset.id;
  if(action==='store'&&shopPatient()){ui.patientPage='shop';closeModal();render();}
  if(action==='category'&&['全部',...shopCategories].includes(id)){shopUi.category=id;render();}
  if(action==='product'){
    const product=state.commerce.products.find(item=>item.id===id);if(!product||shopPatient()&&!product.active)return;
    shopUi.productId=id;
    if(shopPatient()){ui.patientPage='shopProduct';render();}else{modalContext={kind:'shop-preview'};showModal(escapeHtml(product.name),'商品只读预览，不是医护处方',`<p>${escapeHtml(product.description)}</p><p>${escapeHtml(product.usage)}</p><p>${escapeHtml(product.policy)}</p><p>${shopMoney(product.price)} · ${escapeHtml(product.spec)}</p>`);}
  }
  if(action==='edit-product')shopProductForm(id);
  if(action==='save-product')shopSaveProduct();
  if(action==='checkout')shopCheckout(id);
  if(action==='place-order')shopPlaceOrder();
  if(action==='orders'&&shopPatient()){shopUi.orderFilter='全部';shopUi.orderSearch='';ui.patientPage='shopOrders';closeModal();render();}
  if(action==='admin-orders'&&shopAdmin()){shopUi.orderFilter='全部';shopUi.orderSearch='';ui.page='shopOrders';closeModal();render();}
  if(action==='order-filter'&&['全部',...shopStatuses,'售后处理中'].includes(id)){shopUi.orderFilter=id;render();}
  if(action==='order'&&shopOwnOrder(id)){shopUi.orderId=id;if(shopAdmin())ui.page='shopOrder';else ui.patientPage='shopOrder';closeModal();render();}
  if(action==='gateway'&&shopAdmin()){ui.page='shopGateway';closeModal();render();}
  if(action==='save-gateway'&&shopAdmin()){
    const merchant=read('shop-merchant'),enabled=checked('shop-enabled'),wechat=checked('shop-wechat'),alipay=checked('shop-alipay'),outcome=read('shop-outcome'),refundOutcome=read('shop-refund-outcome');
    if(!merchant||merchant.length>60||enabled&&!wechat&&!alipay||![outcome,refundOutcome].every(value=>['success','failure'].includes(value)))return formError('请填写演示商户名称，启用时至少选择一个渠道，并设置有效的模拟结果。');
    state.commerce.gateway={merchant,enabled,wechat,alipay,outcome,refundOutcome};persist();toast('模拟支付配置已保存，无真实支付连接');
  }
  if(['pay','cancel','address','complete','after-sale','appeal','return','ship','review','return-received','refund'].includes(action))shopOperationForm(action,id);
  if(action==='confirm-operation')shopConfirmOperation();
},true);
function shopSearchChange(event){
  if(!['shop-search','shop-order-search'].includes(event.target.id))return;
  shopUi[event.target.id==='shop-search'?'search':'orderSearch']=event.target.value.trim();render();
}
document.addEventListener('change',shopSearchChange);
document.addEventListener('keydown',event=>{if(event.key==='Enter')shopSearchChange(event);});
