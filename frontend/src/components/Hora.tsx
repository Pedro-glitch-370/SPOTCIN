import { useState, useEffect } from 'react';


function HoraAtual() {
  const [agora, setAgora] = useState<Date>(new Date());

  useEffect(() => {
    const intervalo = setInterval(() => {
      setAgora(new Date());
    }, 1000);

    return () => clearInterval(intervalo);
  }, []);

  // 1. Formata a hora para mostrar na tela (ex: 22:19:05)
  /*const horaFormatada = agora.toLocaleTimeString('pt-BR', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });*/

  // 2. Pega apenas a hora como número (0 a 23) para checar o período do dia
  const horaNumero = agora.getHours()

  let saudacao = '';
  let titulo = '';
  if (horaNumero >= 5 && horaNumero < 12) {
    titulo = 'Bom dia!';
    saudacao = 'Qual é a boa de hoje?';
  } else if (horaNumero >= 12 && horaNumero < 18) {
    titulo = 'Boa tarde!';
    saudacao = 'Que som combina com agora?';
  } else {
    titulo = 'Boa noite!';
    saudacao = 'Hora de desacelerar.';
  }

  return (
    <div className='Texto'>
      <h1 className='Titulo'>{titulo}</h1>
      <h2 className='Sub'>{saudacao}</h2>
    </div>

  );
}

export default HoraAtual;