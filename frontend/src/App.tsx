import './App.css';
import HoraAtual from './tsx/Hora';
import Cin from '../src/assets/Cin.png';

function App() {


  return (
    <div className='app'>
      <h1>Spotcin</h1>
      <div className='painel'>
        <img src={Cin} alt="Logo do cin" className='Logo' />
        <HoraAtual></HoraAtual>
      </div>
      
    </div>
  )
}

export default App
