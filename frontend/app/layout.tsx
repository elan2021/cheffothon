import type { Metadata } from 'next';
import './globals.css';
export const metadata: Metadata = {title:'Chef Othon | Cardápio',description:'Bolos por encomenda e doces para adoçar o seu dia.'};
export default function Layout({children}:{children:React.ReactNode}) {return <html lang="pt-BR"><body>{children}</body></html>;}
