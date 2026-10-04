test.beforeEach(async({context})=>{await context.route('https://wa.me/**',route=>route.fulfill({status:200,body:'WhatsApp test'}));});
import {test,expect} from '@playwright/test';
test('pedido completo e layout responsivo',async({page})=>{
 await page.goto('/');await expect(page.getByRole('heading',{name:'Fatia sabor Kinder'})).toBeVisible();
 for(const width of [375,414,768,1024,1440]){await page.setViewportSize({width,height:900});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=window.innerWidth)).toBeTruthy();}
 await page.getByRole('button',{name:'Ver mais sobre Fatia sabor Kinder'}).click();
 await page.getByRole('button',{name:'Adicionar Fatia sabor Kinder à sacola'}).click();
 await expect(page.locator('#sacola')).toHaveCount(0);
 await page.getByRole('button',{name:'Sacola com 1 produtos'}).click();
 await expect(page.locator('#sacola')).toContainText('1 × Fatia sabor Kinder');
 await page.getByRole('button',{name:'Continuar pedido'}).click();await page.getByLabel('Seu nome').fill('Teste Navegador');await page.getByLabel('Telefone com DDD').fill('11999999999');
 await page.getByRole('button',{name:'Finalizar pelo WhatsApp'}).click();await expect(page.getByRole('heading',{name:'Pedido recebido!'})).toBeVisible();
 await page.screenshot({path:'test-results/checkout.png',fullPage:true});await page.getByRole('button',{name:'Voltar ao cardápio'}).click();
 await page.setViewportSize({width:1440,height:1000});await page.screenshot({path:'test-results/desktop.png',fullPage:true});
 await page.setViewportSize({width:375,height:900});await page.screenshot({path:'test-results/mobile.png',fullPage:true});
});
test('quantidade só vai para a sacola ao adicionar',async({page})=>{
 await page.setViewportSize({width:375,height:900});await page.goto('/');
 await expect(page.locator('.product').first().getByRole('button')).toHaveCount(1);
 await page.getByRole('button',{name:'Ver mais sobre Fatia sabor Kinder'}).click();
 await expect(page.getByRole('dialog')).toBeVisible();
 await page.getByRole('button',{name:'Aumentar quantidade de Fatia sabor Kinder'}).click();
 await expect(page.locator('#sacola')).toHaveCount(0);
 await page.getByRole('button',{name:'Adicionar Fatia sabor Kinder à sacola'}).click();
 await page.getByRole('button',{name:'Sacola com 2 produtos'}).click();
 await expect(page.locator('#sacola')).toContainText('2 × Fatia sabor Kinder');
 await page.getByRole('button',{name:'Fechar sacola'}).click();
 await expect(page.getByRole('status')).toContainText('adicionado à sacola');
 await page.reload();await page.getByRole('button',{name:'Sacola com 2 produtos'}).click();await expect(page.locator('#sacola')).toContainText('2 × Fatia sabor Kinder');
});
test('encomenda como orçamento',async({page})=>{await page.goto('/');await page.getByRole('button',{name:'Personalizar meu bolo'}).click();await page.getByLabel('Seu nome').fill('Teste Bolo');await page.getByLabel('Telefone com DDD').fill('11999999999');await page.getByLabel('Data da comemoração').fill('2027-12-20');await page.getByRole('button',{name:'Solicitar orçamento'}).click();await expect(page.getByRole('heading',{name:'Sua ideia chegou!'})).toBeVisible();});




