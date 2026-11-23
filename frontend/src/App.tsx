import './Css/App.css';
import HoraAtual from './components/Hora';
import Navbar from './components/Navbar';
import Cin from '../src/assets/Cin.png';
import Playlist from './components/Playlist';

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
