import './Css/App.css';
import HoraAtual from './tsx/Hora';
import Navbar from './tsx/Navbar';
import Cin from '../src/assets/Cin.png';
import Playlist from './tsx/Playlist';

function App() {


  return (
    <div className='app'>

      <Navbar></Navbar>
      <div className='painel'>
        <img src={Cin} alt="Logo do cin" className='Logo' />
        <HoraAtual></HoraAtual>
        <Playlist></Playlist>
      </div>
      
    </div>
  )
}

export default App
