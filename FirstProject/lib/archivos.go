package lib

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	//"github.com/go-delve/delve/pkg/dwarf/reader"
	"math"
	//"unicode"

	//"github.com/go-delve/delve/pkg/dwarf/reader"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//cuidar la MAYUSCULA y la minuscula

//CONSTCOMANDO const de comandos
type CONSTCOMANDO int

//var d Direction = North
const (
	//MKDISK const
	MKDISK CONSTCOMANDO = iota
	//RMDISK const
	RMDISK
	//FDISK const
	FDISK
	//MOUNT const
	MOUNT
	//UNMOUNT const
	UNMOUNT
	//REP const
	REP
	//EXEC const
	EXEC
	//MKFS const
	MKFS
	//LOGIN const
	LOGIN
	//LOGOUT const
	LOGOUT
	//MKGRP const
	MKGRP
	//RMGRP const
	RMGRP
	//MKUSR const
	MKUSR
	//RMUSR const
	RMUSR
	//CHMOD const
	CHMOD
	//MKFILE const
	MKFILE
	//CAT const
	CAT
	//RM const
	RM
	//EDIT const
	EDIT
	//REN const
	REN
	//MKDIR const
	MKDIR
	//CP const
	CP
	//MV const
	MV
	//FIND const
	FIND
	//CHOWN const
	CHOWN
	//CHGRP const
	CHGRP
	//ERROR const
	ERROR
)

//dt := time.Now()
//dt.Format("01-02-2006 15:04:05")

//MBRStruct estructura MBR
type MBRStruct struct {
	MbrTamanio       int64
	MbrFechaCreacion [25]byte
	MbrDiskSignature int64
	Partition        [4]PartitionStruct
}

//PartitionStruct estructura partition para el MBR
type PartitionStruct struct {
	//PartStatus activa o no? [0 || 1]
	PartStatus byte
	//PartType tipo de la particion, [P || E]
	PartType byte
	//PartFit tipo de ajuste de la particion [B || F || W]
	PartFit byte
	//PartStart en que byte del disco inicia la particion
	PartStart int64
	//PartSize tamanio en bytes de la particion
	PartSize int64
	//PartName nombre de la particion
	PartName [16]byte
}

//EBRStruct estructura EBR
type EBRStruct struct {
	//PartStatus esta activa o no
	PartStatus byte
	//PartFit tipo de ajuste B,F,W
	PartFit byte
	//PartStart byte del disco donde inicia la particion
	PartStart int64
	//PartSize tamanio total en bytes de la particion
	PartSize int64
	//PartNext byte donde esta el proximo EBR [-1] si no hay siguiente
	PartNext int64
	//PartName nombre de la particion
	PartName [16]byte
}

//NodoParticion nodo de una particion montada
type NodoParticion struct {
	partition       *PartitionStruct
	ebr             *EBRStruct
	nombre          string
	id              string
	letraDisco      string
	islogeado       bool
	isRoot          bool
	isPartFormatLWH bool
	usuario         [10]byte
	contrasena      [10]byte
}

//NodoDisco nodo de un disco montado
type NodoDisco struct {
	path             string
	letraDisco       string
	listadoParticion *[]*NodoParticion
	contador         int
}

type Libre struct {
	Lstart int
	Lend   int
	Lsize  int
}

//bd unidad
type bdTree struct {
	bdUnidad    *BloqueDeDatosStruct
	posBitmapBD int64
}

//inodo y su lista de bd
type inodoTree struct {
	inodoUnidad    *InodoStruct
	posBitmapInodo int64
	ListaBD        *[]*bdTree
}

//dd y su lista de inodos
type ddTree struct {
	ddUnidad    *DDStruct
	posBitmapDD int64
	ListaInodo  *[]*inodoTree
}

//avd y su lista de dd
type avdTree struct {
	avdUnidad    *AVDStruct
	posBitmapAVD int64
	ListaDD      *[]*ddTree
}

//var nodoDis *NodoDisco
//nodoDis = new(NodoDisco) //==*NodoDisco
//nodoDis.listadoParticion = new([]*NodoParticion)  //==*[]*NodoParticion
//*nodoDis.listadoParticion = make([]*NodoParticion, 0)
//slice assign/pass-arguments is Reference, point to original slice but not array  https://ahuigo.github.io/b/go/10.go-slice#/

//Languages VARIABLES
var Languages map[string]string

//letras lista de letras
var letras map[int]string

//mountList lista de particiones montadas
var listaMontados []*NodoDisco

//contadorDiscos controla el numero de disco a asignar
var contadorDiscos int

//coloresDISK listado de colores para el reporte DISK
var coloresDISK []string

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ESTRUCTURAS SISTEMAS DE ARCHIVOS LWH

//SuperBootStruct contiene informacion general de el sistema de archivos
type SuperBootStruct struct {
	SbNombreHd [50]byte //nombre del disco duro virtual

	SbArbolVirtualCount      int64 //cantidad de estructuras en el arbol virtual de DIRECTORIOS //ARBOL VIRTUAL DIRECTORIOS
	SbDetalleDirectorioCount int64 //cantidad de estructuras en el detalle de DIRECTORIOS //DETALLE DE DIRECTORIOS
	SbInodosCount            int64 //cantidad de Inodos en la tabla de Inodos //TABLA DE INODOS
	SbBloquesCount           int64 //cantidad de bloques de datos //BLOQUES DE DATOS

	SbArbolVirtualFree      int64 //cantidad de estructuras en el arbol de directorios libres // ARBOL VIRTUAL DIRECTORIOS LIBRES
	SbDetalleDirectorioFree int64 //cantidad de estructuras en el detalle de directoriso libres //DETALLE DE DIRECTORIOS LIBRES
	SbInodosFree            int64 //cantidad de inodos en la tabla de inodos en la tabla de inodos libres //TABLA DE INODOS LIBRES
	SbBloquesFree           int64 //cantidad de bloques de datos libres //BLOQUE DE DATOS LIBRES

	SbDateCreacion [25]byte //fecha de creacion del sistema

	SbDateUltimoMontaje [25]byte //ultima fecha de montaje

	SbMontajesCount int64 //cantidad de montajes del sistema LWH

	SbApBitMapArbolDirectorio int64 //apuntador al inicio del bitmap del arbol virtual de directorio //APUNTADOR AL INICIO DEL BITMAP DEL ARBOL VIRTUAL DE DIRECTORIO
	SbApArbolDirectorio       int64 //apuntador al inicio del arbol virtual de directorio

	SbApBitmapDetalleDirectorio int64 //apuntador al inicio del bitmap de detalle de directorio
	SbApDetalleDirectorio       int64 //apuntador al inicio del detalle directorio

	SbApBitmapTablaInodo int64 //apuntador al inicio del bitmap de la tabla de inodos
	SbApTablaInodo       int64 //apuntador al inicio de la tabla de inodos

	SbApBitmapBloques int64 //apuntador al inicio del bitmap de bloques de datos
	SbApBloques       int64 //apuntador al inicio del bloque de datos

	SbApLog int64 //apuntador al inicio del log o bitacora

	SBApSBCopy int64 //apuntador al inicio de la copia del SB

	SbSizeStructArbolDirectorio   int64 //tamanio de una estructura del arbol virtual de directorio
	SbSizeStructDetalleDirectorio int64 //tamanio de la estructura de un detalle de directorio
	SbSizeStructInodo             int64 //tamanio de la estructura de un inodo
	SbSizeStructBloque            int64 //tamanio de la estructura de un bloque de datos

	SbFirstFreeBitArbolDirectorio   int64 //primer bit libre en el bitmap arbol de directorio
	SbFirstFreeBitDetalleDirectorio int64 //primer bit libre en el bitmap detalle de directorio
	SbFirstFreeBitTablaInodo        int64 //primer bit en el bitmap de inodo
	SbFirstFreeBitBloques           int64 //primer bit libre en el bitmap de bloques de datos

	SbMagicNum int64 //numero de carnet del estudiante
}

//AVDStruct estructura par aun arbol virtual de directorios
type AVDStruct struct {
	AVDFechaCreacion            [25]byte
	AVDNombreDirectorio         [50]byte
	AVDApArraySubdirectorios    [6]int64 //arreglo de apuntadores directos a sub-directorios
	AVDApDetalleDirectorio      int64    //un apuntador a un detalle de directorio. Este solo se utilizara en el primer directorio
	AVDApArbolVirtualDirectorio int64    //un apuntador a otro mismo tipo de estructura por si se usan los 6 apuntadores del arreglo de sub-directorios para que puedan seguir creciendo los subdirectorios
	AVDProper                   [50]byte //id del propietario de la carpeta, el que se a generado al momento de crear usuarios. Si el usuario ha sido eliminado, la carpeta pasa a ser propiedad de ROOT
}

//DDinfoStruct nodo del arreglo de la informacion para un Detalle de directorio
type DDinfoStruct struct {
	DDFileNombre           [16]byte //nombre del archivo
	DDFileApInodo          int64    //apuntador al inodo
	DDFileDateCreacion     [25]byte //fecha de creacion del archivo
	DDFileDateModificacion [25]byte //fecha de modificacion del archivo
}

//DDStruct estructura para el detalle de directorio
type DDStruct struct {
	DDArrayFile           [5]DDinfoStruct //arreglo de estructura de tamanio 5
	DDApDetalleDirectorio int64           //Si se quiere ingresar un sexto archivo y no hay espacio en esta estructura de detalle directorio, apunta a otra estructura de detalle de directorio.
}

//InodoStruct tabla que contiene las estructuras para el manejo de archivos de directorios, donde los archivos son manejados por la tabla i-nodos
type InodoStruct struct {
	ICountInodo            int64    //numero de i-nodo
	ISizeArchivo           int64    //tamanio del archivo
	ICountBloquesAsignados int64    //numero de bloques asignados
	IArrayBloques          [4]int64 //arreglo de 4 aputandores a bloques de datos para guardar el archivo
	IApIndirecto           int64    //un apuntador indirecto por si el archivo ocupa mas de 4 bloques de datos, para el manejo de archivos de tamanio "grande"
	IIdProper              [16]byte //identificador del propietario del archivo
}

//BloqueDeDatos contiene la informacion de un archivo
type BloqueDeDatosStruct struct {
	DbData [25]byte //contien la informacion del archivo
}

//Bitacora maneja todas las transacciones que realiza el sistema de archivos(respaldo para recuperacion)
type BitacoraStruct struct {
	LogTipoOperacion [16]byte //tipo de operacion a realizarse
	LogTipo          int64    //si es archivo(0), si es directorio(1)
	LogNombre        [16]byte //nombre del archivo, o directorio
	LogContenido     [25]byte //si hay datos contenidos
	LogFecha         [25]byte //fecha de transaccion
}

//CopiaSuperBootStruct COPIA contiene informacion general de el sistema de archivos
type CopiaSuperBootStruct struct {
	SbNombreHd [50]byte //nombre del disco duro virtual

	SbArbolVirtualCount      int64 //cantidad de estructuras en el arbol virtual de DIRECTORIOS //ARBOL VIRTUAL DIRECTORIOS
	SbDetalleDirectorioCount int64 //cantidad de estructuras en el detalle de DIRECTORIOS //DETALLE DE DIRECTORIOS

	SbInodosCount  int64 //cantidad de Inodos en la tabla de Inodos //TABLA DE INODOS
	SbBloquesCount int64 //cantidad de bloques de datos //BLOQUES DE DATOS

	SbArbolVirtualFree      int64 //cantidad de estructuras en el arbol de directorios libres // ARBOL VIRTUAL DIRECTORIOS LIBRES
	SbDetalleDirectorioFree int64 //cantidad de estructuras en el detalle de directoriso libres //DETALLE DE DIRECTORIOS LIBRES

	SbInodosFree  int64 //cantidad de inodos en la tabla de inodos en la tabla de inodos libres //TABLA DE INODOS LIBRES
	SbBloquesFree int64 //cantidad de bloques de datos libres //BLOQUE DE DATOS LIBRES

	SbDateCreacion [25]byte //fecha de creacion del sistema

	SbDateUltimoMontaje [25]byte //ultima fecha de montaje

	SbMontajesCount int64 //cantidad de montajes del sistema LWH

	SbApBitMapArbolDirectorio int64 //apuntador al inicio del bitmap del arbol virtual de directorio //APUNTADOR AL INICIO DEL BITMAP DEL ARBOL VIRTUAL DE DIRECTORIO
	SbApArbolDirectorio       int64 //apuntador al inicio del arbol virtual de directorio

	SbApBitmapDetalleDirectorio int64 //apuntador al inicio del bitmap de detalle de directorio
	SbApDetalleDirectorio       int64 //apuntador al inicio del detalle directorio

	SbApBitmapTablaInodo int64 //apuntador al inicio del bitmap de la tabla de inodos
	SbApTablaInodo       int64 //apuntador al inicio de la tabla de inodos

	SbApBitmapBloques int64 //apuntador al inicio del bitmap de bloques de datos
	SbApBloques       int64 //apuntador al inicio del bloque de datos

	SbApLog int64 //apuntador al inicio del log o bitacora

	SBApSBCopy int64 //apuntador al inicio de la copia del SB

	SbSizeStructArbolDirectorio   int64 //tamanio de una estructura del arbol virtual de directorio
	SbSizeStructDetalleDirectorio int64 //tamanio de la estructura de un detalle de directorio
	SbSizeStructInodo             int64 //tamanio de la estructura de un inodo
	SbSizeStructBloque            int64 //tamanio de la estructura de un bloque de datos

	SbFirstFreeBitArbolDirectorio   int64 //primer bit libre en el bitmap arbol de directorio
	SbFirstFreeBitDetalleDirectorio int64 //primer bit libre en el bitmap detalle de directorio
	SbFirstFreeBitTablaInodo        int64 //primer bit en el bitmap de inodo
	SbFirstFreeBitBloques           int64 //primer bit libre en el bitmap de bloques de datos
	SbMagicNum                      int64 //numero de carnet del estudiante
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

var cantidadFormateosLWH int64

//init initialization code here
func init() {
	Languages = make(map[string]string)
	Languages["cs"] = "C #"
	Languages["js"] = "JavaScript"
	Languages["rb"] = "Ruby"
	Languages["go"] = "Golang"

	listaMontados = make([]*NodoDisco, 0)
	letras = make(map[int]string)
	letras[0] = "a"
	letras[1] = "b"
	letras[2] = "c"
	letras[3] = "d"
	letras[4] = "e"
	letras[5] = "f"
	letras[6] = "g"
	letras[7] = "h"
	letras[8] = "i"
	letras[9] = "j"
	letras[10] = "k"
	letras[11] = "l"
	letras[12] = "m"
	letras[13] = "n"
	letras[14] = "o"
	letras[15] = "p"
	letras[16] = "q"
	letras[17] = "r"
	letras[18] = "s"
	letras[19] = "t"
	letras[20] = "u"
	letras[21] = "v"
	letras[22] = "w"
	letras[23] = "x"
	letras[24] = "y"
	letras[25] = "z"
	//fmt.Println("ME ESTOY INICIALIZANDO .......................................................................")

	contadorDiscos = 0

	coloresDISK = make([]string, 0)
	coloresDISK = append(coloresDISK, "#A5C1CF", "#11AAF7", "#F78711", "#FF5733", "#1FA853", "#11AAF7", "#00FFFF", "#A8E828", "#FFFF00")

	cantidadFormateosLWH = 0
}

//CadenaConsola pide una linea de consola hasta el \n
func CadenaConsola() string {
	reader := bufio.NewReader(os.Stdin)
	entrada, _ := reader.ReadString('\n')
	nombre := strings.TrimRight(entrada, "\r\n\t")
	return nombre
}

//Separar verifica si viene \* y separa por espacios
func Separar(cadena string) {

	if strings.Contains(cadena, "#") {
		return
	}

	if cadena == "pause" || cadena == "Pause" {
		fmt.Println("Ingresa un caracter para continuar...")
		CadenaConsola()
		return
	}

	mapa := make(map[string]string) //TODO:PUEDE ESTAR nil
	var commandString CONSTCOMANDO

	if strings.Contains(cadena, "\\*") {
		cadena = strings.ReplaceAll(cadena, "\\*", " ")
	}

	mapa["INSTRUCCION"] = cadena

	segmentos := strings.Split(cadena, " -")

	for i, val := range segmentos {
		porciones := strings.Split(val, "->")

		porciones[0] = strings.ToUpper(porciones[0]) //MKDISK//SIZE//16//PATH//"/home/mis discos/"//NAME->Disco_4.dsk

		if i == 0 {
			mapa[porciones[0]] = porciones[0]
			commandString = Cons(mapa[porciones[0]])

			if commandString == -1 {
				PrintError(commandString, "El comando que ingresaste no es correcto")
				return
			}

		} else {

			if elem, ok := mapa[porciones[0]]; ok {
				PrintAviso(commandString, "El parametro "+porciones[0]+" ya existe, se tomara el valor "+elem+" ingresado")
				continue
			}
			if porciones[0] != "P" {
				mapa[porciones[0]] = porciones[1]
			} else {
				mapa[porciones[0]] = porciones[0]
			}
		}
	}

	if _, found := mapa["PATH"]; found {
		mapa["PATH"] = strings.ReplaceAll(mapa["PATH"], "\"", "")
	}
	if _, found := mapa["RUTA"]; found {
		mapa["RUTA"] = strings.ReplaceAll(mapa["RUTA"], "\"", "")
	}
	if _, found := mapa["NAME"]; found {
		mapa["NAME"] = strings.ReplaceAll(mapa["NAME"], "\"", "")
	}
	if _, found := mapa["DEST"]; found {
		mapa["DEST"] = strings.ReplaceAll(mapa["DEST"], "\"", "")
	}

	fmt.Println(mapa)
	ConstruirComando(commandString, mapa)

}

//ConstruirComando valida parametros y manda a ejecutar un comando
func ConstruirComando(comando CONSTCOMANDO, mapa map[string]string) { //el mapa no tiene posiciones fijos pero si keys fijos

	switch comando {
	case MKDISK:
		if _, ok := mapa["MKDISK"]; !ok {
			PrintError(comando, "El comando mkdisk no esta en la sentencia")
			return
		}
		//VERIFICO-PROCESO-EJECUTO-LISTO PARA PROCEDER [COLUMNA DEL ENUNCIADO]
		//---------------------------------------------------------------------------------------SIZE
		resultA := ParametroSize(comando, mapa)
		//---------------------------------------------------------------------------------------PATH
		resultB := ParametroPathMkdisk(comando, mapa)
		//---------------------------------------------------------------------------------------NAME
		resultC := ParametroNameMkdisk(comando, mapa)
		//---------------------------------------------------------------------------------------UNIT
		resultD := ParametroUnit(comando, mapa)
		//---------------------------------------------------------------------------------------------MKDISK

		if resultA > 0 && resultB == 0 && resultC == 0 && resultD == 0 {
			ComandoMkdisk(comando, mapa, resultA)
			return
		} else {
			PrintError(comando, "Por la existencia de algun error no se procede a ejecutar el comando [size:"+strconv.Itoa(resultA)+",path:"+strconv.Itoa(resultB)+",name:"+strconv.Itoa(resultC)+",unit"+strconv.Itoa(resultD)+"]")
			return
		}

	case RMDISK:
		if _, ok := mapa["RMDISK"]; !ok {
			PrintError(comando, "El comando rmdisk no esta en la sentencia")
			return
		}
		//---------------------------------------------------------------------------------------PATH
		//---------------------------------------------------------------------------------------------RMDISK
		ComandoRmdisk(comando, mapa)
		return

	case FDISK:
		//TODO: QUIZA EVALUE VALORES QUE NO TOMARA EN CUENTA PARA DELETE, ADD, CREAR
		if _, ok := mapa["FDISK"]; !ok {
			PrintError(comando, "El comando fdisk no esta en la sentencia")
			return
		}
		//VERIFICO-|PROCESO-EJECUTO|-LISTO PARA PROCEDER [COLUMNA DEL ENUNCIADO]
		//---------------------------------------------------------------------------------------SIZE
		resultSize := ParametroSize(comando, mapa)
		//---------------------------------------------------------------------------------------UNIT
		resultUnit := ParametroUnitFdisk(comando, mapa)
		//---------------------------------------------------------------------------------------PATH
		resultPath := ParametroPathFdisk(comando, mapa)
		//---------------------------------------------------------------------------------------TYPE
		resultType := ParametroTypeFdisk(comando, mapa)
		//---------------------------------------------------------------------------------------FIT
		resultFit := ParametroFitFdisk(comando, mapa)
		//---------------------------------------------------------------------------------------NAME
		resultName := ParametroNameFdisk(comando, mapa)
		//---------------------------------------------------------------------------------------DELETE
		resultDeleteBool, resultDeleteInt := ParametroDeleteFdisk(comando, mapa)
		//---------------------------------------------------------------------------------------ADD
		resultAddBool, resultAddInt := ParametroAddFdisk(comando, mapa)
		//-----------------------------------------------------------------------------------------------------------------

		if !resultDeleteBool && !resultAddBool { //sino viene ninguno de los dos

			if resultSize > 0 && resultUnit == 0 && resultPath == 0 && resultType == 0 && resultFit == 0 && resultName == 0 {
				ComandoFdisk(comando, mapa, resultSize)
				return
			} else {
				PrintError(comando, "Por la existencia de algun error no se procede a ejecutar el comando [size:"+strconv.Itoa(resultSize)+", unit:"+strconv.Itoa(resultUnit)+", path:"+strconv.Itoa(resultPath)+", type"+strconv.Itoa(resultType)+", fit"+strconv.Itoa(resultFit)+", name"+strconv.Itoa(resultName)+"]")
				return
			}

		} else {

			if resultDeleteBool && resultAddBool {

				PrintError(comando, "Error al ejecutar Fdisk, pueden venir dos parametros no permitos en la misma sentencia [resultDeleteBool:"+strconv.FormatBool(resultDeleteBool)+", resultAddBool:"+strconv.FormatBool(resultAddBool)+"]")
				return

			} else if resultDeleteBool {

				if resultDeleteInt == 0 {

					if resultPath == 0 && resultName == 0 { //tenemos path y name bien
						EjecutarDelete(mapa["PATH"], mapa["NAME"], mapa["DELETE"], comando)
						return
					}
					PrintError(comando, "Existe algun error con algun parametro obligatorio para usar delete [resultPath:"+strconv.Itoa(resultPath)+", resultName:"+strconv.Itoa(resultName)+"]")
					return

				}
				//==-1
				PrintError(comando, "Viene el parametro delete pero con errores [resultDeleteInt:"+strconv.Itoa(resultDeleteInt)+"]")
				return

			} else if resultAddBool {

				if resultAddInt != 0 { //==valor

					if resultUnit == 0 && resultPath == 0 && resultName == 0 {
						//tomara unit para saber cuanto quitar b,k,m
						EjecutarAdd(mapa["PATH"], mapa["NAME"], resultAddInt, mapa["UNIT"], comando)
						return
						//TODO EVALUACION
					}
					PrintError(comando, "Existe algun error con algun parametro obligatorio para usar add [resultUnit:"+strconv.Itoa(resultUnit)+", resultPath:"+strconv.Itoa(resultPath)+", resultName:"+strconv.Itoa(resultName)+"]")
					return
				}
				//==0
				PrintAviso(comando, "Estas agregando/quitando un tamanio == [0], no se procede a hacer algo, todo quedaria igual")
				return

			} else {
				PrintError(comando, "Error al ejecutar Fdisk, verifica los parametos ||crear una particion||delete||add|| [resultDeleteBool:"+strconv.FormatBool(resultDeleteBool)+", resultAddBool:"+strconv.FormatBool(resultAddBool)+"]")
				return
			}

		}

	case MOUNT:
		if _, ok := mapa["MOUNT"]; !ok {
			PrintError(comando, "El comando mount no esta en la sentencia")
			return
		}

		if len(mapa) == 2 {
			PrintAviso(comando, "Desplegando la informacion de las particiones montadas en memoria:")
			if len(listaMontados) != 0 {
				//TODO: HACER QUE RETORNE UN BOOL PARA SALIR LUEGO DE ESTO, VER PRIMERO SI ES NECESARIO, AL PARECER NO, PERO MIRAR PRIMERO
				desplegarMount(&listaMontados)
			} else {
				PrintAviso(comando, "No hay particiones montadas aun")
			}
			return
		}
		//---------------------------------------------------------------------------------------PATH
		resultPath := ParametroPathMount(comando, mapa)
		//---------------------------------------------------------------------------------------NAME
		resultName := ParametroNameFdisk(comando, mapa)
		//---------------------------------------------------------------------------------------------MOUNT

		if resultPath == 0 && resultName == 0 {
			ComandoMount(comando, mapa, &listaMontados)
			fmt.Println("LISTADO DE DISCOS MONTADOS:")
			fmt.Println(listaMontados)
		} else {
			PrintError(ERROR, "Por la existencia de algun error no se procede a ejecutar el comando Mount [Path:"+strconv.Itoa(resultPath)+", Name:"+strconv.Itoa(resultName)+"]")
			return
		}

	case UNMOUNT:
		if _, ok := mapa["UNMOUNT"]; !ok {
			PrintError(comando, "El comando unmount no esta en la sentencia")
			return
		}
		if len(mapa) == 2 {
			PrintError(ERROR, "Faltan parametros para ejecutar")
			return
		}
		//---------------------------------------------------------------------------------------IDn
		//resultUnMount := ComandoUnmount(comando, mapa, &listaMontados)
		ComandoUnmount(comando, mapa, &listaMontados)
		//desmontar la particion del arreglo
		//extraerla del arreglo
		//con la informacion de la extraida *EBR *Partition
		//ir a grabarla al disco y/o al MBR [P|L] //si es logica y es la primera particion ir al primer EBR a la Extendida pareciera que no
		//verificar que modifique los valores reales

	case REP:
		if _, ok := mapa["REP"]; !ok {
			PrintError(comando, "El comando rept no esta en la sentencia")
			return
		}
		//---------------------------------------------------------------------------------------NOMBRE
		resultBoolNombre, resultNombre := ParametroNombreRep(comando, mapa)
		//---------------------------------------------------------------------------------------PATH
		resultBoolPath, resultNameFile := ParametroPathRep(comando, mapa)
		//---------------------------------------------------------------------------------------RUTA
		resultRuta := ParametroRutaRep(comando, mapa)
		//---------------------------------------------------------------------------------------ID
		var nodoDis *NodoDisco
		var indexDisc = -1
		var nodoPart *NodoParticion
		var indexPart = -1

		nodoDis, indexDisc, nodoPart, indexPart = ParametroIDRep(comando, mapa, &listaMontados)
		//TODO: IMPRIMIR Y VERIFICAR QUE SI ME TRAE LA INFORMACION CORRECTAMENTE Y LA SINTAXIS NO ME AFECTA
		//---------------------------------------------------------------------------------------RUTA
		//TODO: POR HACER, CUANDO YA TENGAMOS PARA EL REPORTE file Y ls
		//---------------------------------------------------------------------------------------REP

		if nodoDis != nil && nodoPart != nil && resultBoolNombre && resultBoolPath {
			ComandoRep(resultNombre, resultNameFile, nodoDis, nodoPart, indexDisc, indexPart, comando, resultRuta, mapa["PATH"])
		} else {
			//TODO: FALTAN MOSTRAR ALGUNOS PARAMETROS
			PrintError(ERROR, "Por la existencia de algun inconveniente no se procede a ejecutar el comando Rep [Result Nombre: "+strconv.FormatBool(resultBoolNombre)+", Result Path: "+strconv.FormatBool(resultBoolPath)+", Result Disco: "+strconv.Itoa(indexDisc)+", Result Particion: "+strconv.Itoa(indexPart)+"]")
			return
		}

	case EXEC:
		if _, ok := mapa["EXEC"]; !ok {
			PrintError(comando, "El comando exec no esta en la sentencia")
			return
		}

		if boo, arr := ParametroPathExec(comando, mapa); (boo) && arr != nil {
			ComandoExec(arr)
		} else {
			PrintError(comando, "Error al ejecutar el comando Exec ["+strconv.FormatBool(boo)+"] quizas el arreglo este nil")
			return
		}

	case MKFS:
		if _, ok := mapa["MKFS"]; !ok {
			PrintError(comando, "El comando mkfs no esta en la sentencia")
			return
		}

		//---------------------------------------------------------------------------------------ID
		var nodoDis *NodoDisco
		var indexDisc = -1
		var nodoPart *NodoParticion
		var indexPart = -1
		nodoDis, indexDisc, nodoPart, indexPart = ParametroIDRep(comando, mapa, &listaMontados)
		//---------------------------------------------------------------------------------------TIPO
		resultTipo := ParametroTipo(comando, mapa)
		//---------------------------------------------------------------------------------------ADD/UNIT
		resultAddBool, resultAddInt := ParametroAddFdisk(comando, mapa) //TODO: QUIZAS AQUI VALIDAR ESTE PARAMETRO
		resultUnit := ParametroUnitFdisk(comando, mapa)
		//---------------------------------------------------------------------------------------TAMANIO_TOTAL
		tamanioFinal := TamanioTotal(comando, resultAddInt, mapa["UNIT"])
		//---------------------------------------------------------------------------------------MKFS

		if !resultAddBool { //no viene add
			if nodoDis != nil && nodoPart != nil && resultTipo == 0 {
				//TODO:YA FORMATEO LA CARPETA RAIZ
				//TODO:YA CREO CARPETAS RECURSIVAMENTE EN CUALQUIER CASO
				//TODO:YA GRAFICO, bitmap(de cualquiera), ....directorio,
				ComandoMKFS(nodoDis, nodoPart, mapa["TIPO"], "Eduardo", comando, &cantidadFormateosLWH)
			} else {
				PrintError(ERROR, "Por la existencia de algun inconveniente no se procede a ejecutar el comando Mkfs [Result Disco: "+strconv.Itoa(indexDisc)+", Result Particion: "+strconv.Itoa(indexPart)+", Result Tipo: "+strconv.Itoa(resultTipo)+"]")
				return
			}

		} else if resultAddBool { //viene add
			if nodoDis != nil && nodoPart != nil && resultTipo == 0 && resultAddBool && resultUnit == 0 {
				if tamanioFinal != -1 {
					//TODO: SI SE HACE QUE ME RETORNE UN BOOL PARA SABER EL RESULTADO
					//ComandoMKFSAdd(nodoDis, nodoPart, mapa["TIPO"], tamanioFinal, comando)
				} else {
					PrintError(ERROR, "Error al calcular el tamanio final para el parametro add")
					return
				}
			} else {
				PrintError(ERROR, "Por la existencia de algun inconveniente no se procede a ejecutar el comando Mkfs [Result Disco: "+strconv.Itoa(indexDisc)+", Result Particion: "+strconv.Itoa(indexPart)+", Result Tipo: "+strconv.Itoa(resultTipo)+", Result Add Bool: "+strconv.FormatBool(resultAddBool)+", Result Unit: "+strconv.Itoa(resultUnit)+"]")
				return
			}
		} else {
			PrintError(ERROR, "Error al ejecutar el comando mkfs")
			return
		}

	case MKDIR:
		if _, ok := mapa["MKDIR"]; !ok {
			PrintError(comando, "El comando mkdir no esta en la sentencia")
			return
		}
		//---------------------------------------------------------------------------------------ID
		var nodoDis *NodoDisco
		var indexDisc = -1
		var nodoPart *NodoParticion
		var indexPart = -1
		nodoDis, indexDisc, nodoPart, indexPart = ParametroIDRep(comando, mapa, &listaMontados)
		//---------------------------------------------------------------------------------------PATH
		resultPath := ParametroPathMkdir(comando, mapa)

		if _, found := mapa["P"]; !found { //sino existe P //TODO: VERIFICAR QUE LE QUITE !
			//sino existe P
			//no existe las carpetas error, porque no viene el otro parametro
			if nodoDis != nil && nodoPart != nil && resultPath == 0 {

				//if ValidacionParticion(nodoDis, nodoPart, comando){
				if valBool, valFit := partFitParticion(nodoPart); valBool {

					MkdirFinal(nodoDis.path, mapa["PATH"], "Eduardo", valFit, false, comando, nodoDis, nodoPart)
					PrintAviso(comando, "fin de ejecucion de mkdir final")
					return

				} else {
					PrintError(ERROR, "Error al intentar averiguar que fit tiene la particion")
					return
				}

				//}else{
				//	PrintError(ERROR, "No cumple con ciertas validaciones, para continuar")
				//	return
				//}
			} else {
				PrintError(ERROR, "Por la existencia de un error no se procede a ejecutar el comando [Nodo Disco: "+strconv.Itoa(indexDisc)+", Nodo Part: "+strconv.Itoa(indexPart)+", Path: "+strconv.Itoa(resultPath)+"]")
				return
			}
		} else { //existe P
			//if no existen las carpetas padres se crean
			if nodoDis != nil && nodoPart != nil && resultPath == 0 {

				//if ValidacionParticion(nodoDis, nodoPart, comando){
				if valBool, valFit := partFitParticion(nodoPart); valBool {

					MkdirFinal(nodoDis.path, mapa["PATH"], "Eduardo", valFit, true, comando, nodoDis, nodoPart)
					PrintAviso(comando, "fin de ejecucion de mkdir final")
					return

				} else {
					PrintError(ERROR, "Error al intentar averiguar que fit tiene la particion")
					return
				}

				//}else{
				//	PrintError(ERROR, "No cumple con ciertas validaciones, para continuar")
				//	return
				//}

			} else {
				PrintError(ERROR, "Por la existencia de un error no se procede a ejecutar el comando [Nodo Disco: "+strconv.Itoa(indexDisc)+", Nodo Part: "+strconv.Itoa(indexPart)+", Path: "+strconv.Itoa(resultPath)+"]")
				return
			}
		}

	case MKFILE:
		if _, ok := mapa["MKFILE"]; !ok {
			PrintError(comando, "El comando mkfile no esta en la sentencia")
			return
		}
		//---------------------------------------------------------------------------------------ID
		var nodoDis *NodoDisco
		var indexDisc = -1
		var nodoPart *NodoParticion
		var indexPart = -1
		nodoDis, indexDisc, nodoPart, indexPart = ParametroIDRep(comando, mapa, &listaMontados)
		//---------------------------------------------------------------------------------------PATH
		resultPath := ParametroPathMkdir(comando, mapa)
		//---------------------------------------------------------------------------------------ISP
		resultisP := ParametroisP(comando, mapa)
		//---------------------------------------------------------------------------------------SIZE
		resultSizeBool, resultSizeVal := ParametroSizeMkfile(comando, mapa) //puede ser == 0
		//---------------------------------------------------------------------------------------CONT
		resultisCont := ParametroCont(comando, mapa) //puede ser == ""

		if nodoDis != nil && nodoPart != nil && resultPath == 0 { //obligatorios

			arrDir := ArrDir(mapa["PATH"])
			if arrDir != nil { //archivo afuera de la raiz

				var sb *SuperBootStruct
				sb = new(SuperBootStruct)
				var sbCopia *SuperBootStruct
				sbCopia = new(SuperBootStruct)

				if ObtenerSBySBCopia(nodoDis, nodoPart, sb, sbCopia, comando) {

					if valBool, valFit := partFitParticion(nodoPart); valBool {

						size, contenido := MkfileOpcionales(resultSizeBool, resultSizeVal, resultisCont, mapa["CONT"])
						dir, _ := filepath.Split(mapa["PATH"])
						if comandoMkfile(nodoDis.path, nodoPart, sb, sbCopia, arrDir, dir, resultisP, size, contenido, comando, "Eduardo", valFit) { //size==0 || contenido=="", pero siempre iguales y el size manda
							PrintAviso(comando, "se creo el archivo exitosamente")
							return
						} else {
							PrintError(ERROR, "No se pudo crear el archivo por algun inconveniente")
							return
						}

					} else {
						PrintError(ERROR, "Error al intentar averiguar que fit tiene la particion")
						return
					}

				} else {
					PrintError(ERROR, "Existio un error al momento de extraer el SB y el SB Copia de la particion")
					return
				}

			} else {
				PrintError(ERROR, "No puedes crear un archivo afuera de la raiz, verifica la ruta porfavor")
				return
			}
		} else {
			PrintError(ERROR, "Por la existencia de un error no se procede a ejecutar el comando [Nodo Disco: "+strconv.Itoa(indexDisc)+", Nodo Part: "+strconv.Itoa(indexPart)+", Path: "+strconv.Itoa(resultPath)+"]")
			return
		}

	case CAT:
		if _, ok := mapa["CAT"]; !ok {
			PrintError(comando, "El comando cat no esta en la sentencia")
			return
		}
		//---------------------------------------------------------------------------------------ID
		var nodoDis *NodoDisco
		var indexDisc = -1
		var nodoPart *NodoParticion
		var indexPart = -1
		nodoDis, indexDisc, nodoPart, indexPart = ParametroIDRep(comando, mapa, &listaMontados)
		//---------------------------------------------------------------------------------------IDn
		if nodoDis != nil && nodoPart != nil {
			ComandoCat(nodoDis, nodoPart, mapa, comando)
		} else {
			PrintError(ERROR, "Por algun inconveniente no se puede ejecutar el comando [Disco: "+strconv.Itoa(indexDisc)+", Particion: "+strconv.Itoa(indexPart)+"]")
		}

	case RM:
		//TODO: REVISAR PORQUE NO ME ELIMINO LA CARPETA docs
		if _, ok := mapa["RM"]; !ok {
			PrintError(comando, "El comando rm no esta en la sentencia")
			return
		}
		//---------------------------------------------------------------------------------------ID
		var nodoDis *NodoDisco
		var indexDisc = -1
		var nodoPart *NodoParticion
		var indexPart = -1
		nodoDis, indexDisc, nodoPart, indexPart = ParametroIDRep(comando, mapa, &listaMontados)
		if nodoDis != nil && nodoPart != nil {
			//TODO: UN CONDICIONAL PARA EL PARAMETRO PATH Y EL OTRO? COMO PREVENCION
			ComandoRm(nodoDis, nodoPart, mapa["PATH"], comando)
			return

		} else {
			PrintError(ERROR, "Por algun inconveniente no se puede ejecutar el comando rm [Disco: "+strconv.Itoa(indexDisc)+", Particion: "+strconv.Itoa(indexPart)+"]")
			return
		}

	case REN:
		if _, ok := mapa["REN"]; !ok {
			PrintError(comando, "El comando ren no esta en la sentencia")
			return
		}
		//---------------------------------------------------------------------------------------ID
		var nodoDis *NodoDisco
		var indexDisc = -1
		var nodoPart *NodoParticion
		var indexPart = -1
		nodoDis, indexDisc, nodoPart, indexPart = ParametroIDRep(comando, mapa, &listaMontados)
		if nodoDis != nil && nodoPart != nil {
			//TODO: UN CONDICIONAL PARA EL PARAMETRO PATH Y EL OTRO? COMO PREVENCION
			ComandoRen(nodoDis, nodoPart, mapa["PATH"], mapa["NAME"], comando)
			return
		} else {
			PrintError(ERROR, "Por algun inconveniente no se puede ejecutar el comando ren [Disco: "+strconv.Itoa(indexDisc)+", Particion: "+strconv.Itoa(indexPart)+"]")
			return
		}

	case FIND:
		if _, ok := mapa["FIND"]; !ok {
			PrintError(comando, "El comando find no esta en la sentencia")
			return
		}
		//---------------------------------------------------------------------------------------ID
		var nodoDis *NodoDisco
		var indexDisc = -1
		var nodoPart *NodoParticion
		var indexPart = -1
		nodoDis, indexDisc, nodoPart, indexPart = ParametroIDRep(comando, mapa, &listaMontados)
		if nodoDis != nil && nodoPart != nil {
			//TODO: UN CONDICIONAL PARA EL PARAMETRO PATH Y EL OTRO? COMO PREVENCION
			ComandoFind(nodoDis, nodoPart, mapa, comando)
		} else {
			PrintError(ERROR, "Por algun inconveniente no se puede ejecutar el comando find [Disco: "+strconv.Itoa(indexDisc)+", Particion: "+strconv.Itoa(indexPart)+"]")
		}

	case EDIT:
		if _, ok := mapa["EDIT"]; !ok {
			PrintError(comando, "El comando edit no esta en la sentencia")
			return
		}
		//---------------------------------------------------------------------------------------ID
		var nodoDis *NodoDisco
		var indexDisc = -1
		var nodoPart *NodoParticion
		var indexPart = -1
		nodoDis, indexDisc, nodoPart, indexPart = ParametroIDRep(comando, mapa, &listaMontados)
		//---------------------------------------------------------------------------------------SIZE
		resultSizeBool, resultSizeVal := ParametroSizeMkfile(comando, mapa) //puede ser == 0
		//---------------------------------------------------------------------------------------CONT
		resultisCont := ParametroCont(comando, mapa) //puede ser == ""

		if nodoDis != nil && nodoPart != nil {

			if resultSizeBool || resultisCont {
				ComandoEdit(nodoDis, nodoPart, mapa["PATH"], resultSizeVal, mapa["CONT"], mapa["ID"], comando)
				return
			} else {
				PrintError(ERROR, "Debe de venir almenos alguno de los parametros [SIZE: "+strconv.FormatBool(resultSizeBool)+", CONT: "+strconv.FormatBool(resultisCont)+"]")
				return
			}

		} else {
			PrintError(ERROR, "Por algun inconveniente no se puede ejecutar el comando edit [Disco: "+strconv.Itoa(indexDisc)+", Particion: "+strconv.Itoa(indexPart)+"]")
			return
		}

	case CP:
		if _, ok := mapa["CP"]; !ok {
			PrintError(comando, "El comando cp no esta en la sentencia")
			return
		}
		//---------------------------------------------------------------------------------------ID
		var nodoDis *NodoDisco
		var indexDisc = -1
		var nodoPart *NodoParticion
		var indexPart = -1
		nodoDis, indexDisc, nodoPart, indexPart = ParametroIDRep(comando, mapa, &listaMontados)
		if nodoDis != nil && nodoPart != nil {
			ComandoCp(nodoDis, nodoPart, mapa["PATH"], mapa["DEST"], mapa["ID"], comando)
			return
		} else {
			PrintError(ERROR, "Por algun inconveniente no se puede ejecutar el comando cp [Disco: "+strconv.Itoa(indexDisc)+", Particion: "+strconv.Itoa(indexPart)+"]")
		}

	case MV:
		if _, ok := mapa["MV"]; !ok {
			PrintError(comando, "El comando mv no esta en la sentencia")
			return
		}
		//---------------------------------------------------------------------------------------ID
		var nodoDis *NodoDisco
		var indexDisc = -1
		var nodoPart *NodoParticion
		var indexPart = -1
		nodoDis, indexDisc, nodoPart, indexPart = ParametroIDRep(comando, mapa, &listaMontados)
		if nodoDis != nil && nodoPart != nil {
			ComandoMV(nodoDis, nodoPart, mapa["PATH"], mapa["DEST"], mapa["ID"], comando)
			return
		} else {
			PrintError(ERROR, "Por algun inconveniente no se puede ejecutar el comando mv [Disco: "+strconv.Itoa(indexDisc)+", Particion: "+strconv.Itoa(indexPart)+"]")
		}

	default:
		PrintError(comando, "El comando no es correcto ["+comando.String()+"]")
		return
	}

}

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func ComandoMV(nodoDisco *NodoDisco, nodoParticion *NodoParticion, rutaFile string, destinoFile string, id string, comando CONSTCOMANDO) {
	rutaFile = strings.ReplaceAll(rutaFile, "\"", "")
	arrDir := ArrDir(rutaFile)
	archivo := ""
	if arrDir != nil {

		isArchivo := false
		if strings.Contains(rutaFile, ".") { //ruta de archivo

			archivo = string(arrDir[len(arrDir)-1][:])
			arrDir = arrDir[:len(arrDir)-1]
			isArchivo = true

		} else { //ruta de carpeta
			PrintAviso(comando, "EL mover para una carpeta no esta implementado, disculpas.")
			PrintError(ERROR, "Al parecer la ruta dada es de una carpeta y no de un archivo")
			return
		}

		if ejecutarMv(nodoDisco, nodoParticion, arrDir, archivo, comando, isArchivo, destinoFile, id) {

			PrintAviso(comando, "Se movio la ruta exitosamente")
			return

		} else {
			PrintError(ERROR, "error al mover la ruta dada")
			return
		}
	} else {
		PrintError(ERROR, "Existe algun error con el parametro ruta")
		return
	}
}

func ejecutarMv(nodoDisco *NodoDisco, nodoParticion *NodoParticion, arrDir [][50]byte, archivo string, comando CONSTCOMANDO, isArchivo bool, destinoFile string, id string) bool {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		listaAVDS := make([]*avdTree, 0) //lista de avdTress
		treeTres(nodoDisco.path, sb, avd, arrDir, archivo, 0, 0, comando, &listaAVDS)

		if len(listaAVDS) == 0 {
			PrintAviso(comando, "la ruta no existe o no existe completa")
			return false
		}

		if !isArchivo { //carpeta

			PrintAviso(comando, "Al parecer la ruta es de una carpeta, no se puede proceder, a falta de implementacion...")
			return false

		} else { //archivo

			encontrado := false

			nombreArchivo := strings.Trim(archivo, "0")
			destinoFile = destinoFile + "/" + nombreArchivo

			contenidoArchivo := ""
			sizeArchivo := int64(0)

			for _, avd := range listaAVDS {
				//limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitMapArbolDirectorio, sb.SbArbolVirtualCount, "AVD", avd.posBitmapAVD, comando)
				for _, dd := range *avd.ListaDD {
					//limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapDetalleDirectorio, sb.SbDetalleDirectorioCount, "DD", dd.posBitmapDD, comando)
					for _, inodo := range *dd.ListaInodo {

						sizeArchivo = inodo.inodoUnidad.ISizeArchivo

						for i := range dd.ddUnidad.DDArrayFile { //desuniendo
							if dd.ddUnidad.DDArrayFile[i].DDFileApInodo == inodo.posBitmapInodo { //en la pos del arreglo, elimino el/los inodos que tengo en mi arreglo
								dd.ddUnidad.DDArrayFile[i].DDFileNombre = [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
								dd.ddUnidad.DDArrayFile[i].DDFileDateModificacion = [25]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
								dd.ddUnidad.DDArrayFile[i].DDFileDateCreacion = [25]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
								dd.ddUnidad.DDArrayFile[i].DDFileApInodo = -1
								GuardarDD(comando, dd.ddUnidad, nodoDisco.path, int(sb.SbApDetalleDirectorio+(dd.posBitmapDD*sb.SbSizeStructDetalleDirectorio)))
							}
						}
						limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapTablaInodo, sb.SbInodosCount, "INODO", inodo.posBitmapInodo, comando)

						for _, bd := range *inodo.ListaBD { //copiando los bd
							contenidoArchivo += string(bytes.Trim(bd.bdUnidad.DbData[:], "0"))
							encontrado = true
							limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapBloques, sb.SbBloquesCount, "BD", bd.posBitmapBD, comando)

						}
					}
				}
			}

			if encontrado {
				Separar("mkfile -size->" + strconv.Itoa(int(sizeArchivo)) + " -path->\"" + destinoFile + "\" -p -id->" + id + " " + "-cont->" + contenidoArchivo + "")
				return true
			} else {
				PrintError(ERROR, "Error al actualizar el DD del archivo")
				return false
			}

		}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return false
	}
}

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
func ComandoCp(nodoDisco *NodoDisco, nodoParticion *NodoParticion, rutaFile string, destinoFile string, id string, comando CONSTCOMANDO) {
	rutaFile = strings.ReplaceAll(rutaFile, "\"", "")
	arrDir := ArrDir(rutaFile)
	archivo := ""
	if arrDir != nil {

		isArchivo := false
		if strings.Contains(rutaFile, ".") { //ruta de archivo

			archivo = string(arrDir[len(arrDir)-1][:])
			arrDir = arrDir[:len(arrDir)-1]
			isArchivo = true

		} else { //ruta de carpeta
			PrintAviso(comando, "EL copiar para una carpeta no esta implementado, disculpas.")
			PrintError(ERROR, "Al parecer la ruta dada es de una carpeta y no de un archivo")
			return
		}

		if ejecutarCP(nodoDisco, nodoParticion, arrDir, archivo, comando, isArchivo, destinoFile, id) {

			PrintAviso(comando, "Se copio la ruta exitosamente")
			return

		} else {
			PrintError(ERROR, "error al copiar la ruta dada")
			return
		}
	} else {
		PrintError(ERROR, "Existe algun error con el parametro ruta")
		return
	}
}

func ejecutarCP(nodoDisco *NodoDisco, nodoParticion *NodoParticion, arrDir [][50]byte, archivo string, comando CONSTCOMANDO, isArchivo bool, destinoFile string, id string) bool {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		listaAVDS := make([]*avdTree, 0) //lista de avdTress
		treeTres(nodoDisco.path, sb, avd, arrDir, archivo, 0, 0, comando, &listaAVDS)

		if len(listaAVDS) == 0 {
			PrintAviso(comando, "la ruta no existe o no existe completa")
			return false
		}

		if !isArchivo { //carpeta

			PrintAviso(comando, "Al parecer la ruta es de una carpeta, no se puede proceder, a falta de implementacion...")
			return false

		} else { //archivo

			encontrado := false

			nombreArchivo := strings.Trim(archivo, "0")
			destinoFile = destinoFile + "/" + nombreArchivo

			contenidoArchivo := ""
			sizeArchivo := int64(0)

			for _, avd := range listaAVDS {
				//limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitMapArbolDirectorio, sb.SbArbolVirtualCount, "AVD", avd.posBitmapAVD, comando)
				for _, dd := range *avd.ListaDD {
					//limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapDetalleDirectorio, sb.SbDetalleDirectorioCount, "DD", dd.posBitmapDD, comando)
					for _, inodo := range *dd.ListaInodo {
						sizeArchivo = inodo.inodoUnidad.ISizeArchivo
						for _, bd := range *inodo.ListaBD { //copiando los bd
							contenidoArchivo += string(bytes.Trim(bd.bdUnidad.DbData[:], "0"))
							encontrado = true
						}
					}
				}
			}

			if encontrado {
				Separar("mkfile -size->" + strconv.Itoa(int(sizeArchivo)) + " -path->\"" + destinoFile + "\" -p -id->" + id + " " + "-cont->" + contenidoArchivo + "")
				return true
			} else {
				PrintError(ERROR, "Error al actualizar el DD del archivo")
				return false
			}

		}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return false
	}
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func ComandoEdit(nodoDisco *NodoDisco, nodoParticion *NodoParticion, rutaFile string, sizeNew int, contenidoNew string, id string, comando CONSTCOMANDO) {
	rutaFile = strings.ReplaceAll(rutaFile, "\"", "")
	arrDir := ArrDir(rutaFile)
	archivo := ""
	if arrDir != nil {

		isArchivo := false
		if strings.Contains(rutaFile, ".") { //ruta de archivo

			archivo = string(arrDir[len(arrDir)-1][:])
			arrDir = arrDir[:len(arrDir)-1]
			isArchivo = true

		} else { //ruta de carpeta
			PrintError(ERROR, "Al parecer la ruta dada es de una carpeta y no de un archivo")
			return
		}

		if ejecutarEdit(nodoDisco, nodoParticion, arrDir, archivo, comando, isArchivo, sizeNew, contenidoNew, rutaFile, id) {

			PrintAviso(comando, "Se edito la ruta exitosamente")
			return

		} else {
			PrintError(ERROR, "error al editar la ruta dada")
			return
		}
	} else {
		PrintError(ERROR, "Existe algun error con el parametro ruta")
		return
	}
}

func ejecutarEdit(nodoDisco *NodoDisco, nodoParticion *NodoParticion, arrDir [][50]byte, archivo string, comando CONSTCOMANDO, isArchivo bool, sizeNew int, contenidoNew string, rutaFile string, id string) bool {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		listaAVDS := make([]*avdTree, 0) //lista de avdTress
		treeTres(nodoDisco.path, sb, avd, arrDir, archivo, 0, 0, comando, &listaAVDS)

		if len(listaAVDS) == 0 {
			PrintAviso(comando, "la ruta no existe o no existe completa")
			return false
		}

		if !isArchivo { //carpeta

			PrintAviso(comando, "Al parecer la ruta es de una carpeta, no se puede proceder")
			return false

		} else { //archivo

			eliminado := false
			for _, avd := range listaAVDS {
				//limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitMapArbolDirectorio, sb.SbArbolVirtualCount, "AVD", avd.posBitmapAVD, comando)
				for _, dd := range *avd.ListaDD {
					//limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapDetalleDirectorio, sb.SbDetalleDirectorioCount, "DD", dd.posBitmapDD, comando)
					for _, inodo := range *dd.ListaInodo {

						for i := range dd.ddUnidad.DDArrayFile { //desuniendo
							if dd.ddUnidad.DDArrayFile[i].DDFileApInodo == inodo.posBitmapInodo { //en la pos del arreglo, elimino el/los inodos que tengo en mi arreglo
								dd.ddUnidad.DDArrayFile[i].DDFileNombre = [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
								dd.ddUnidad.DDArrayFile[i].DDFileDateModificacion = [25]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
								dd.ddUnidad.DDArrayFile[i].DDFileDateCreacion = [25]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
								dd.ddUnidad.DDArrayFile[i].DDFileApInodo = -1
								if GuardarDD(comando, dd.ddUnidad, nodoDisco.path, int(sb.SbApDetalleDirectorio+(dd.posBitmapDD*sb.SbSizeStructDetalleDirectorio))) {
									eliminado = true
								}
							}
						}

						limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapTablaInodo, sb.SbInodosCount, "INODO", inodo.posBitmapInodo, comando)

						for _, bd := range *inodo.ListaBD { //eliminando los bd
							limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapBloques, sb.SbBloquesCount, "BD", bd.posBitmapBD, comando)
						}
					}
				}
			}

			if eliminado {
				Separar("mkfile -size->" + strconv.Itoa(sizeNew) + " -path->\"" + rutaFile + "\" -p -id->" + id + " " + "-cont->" + contenidoNew + "")
				return true
			} else {
				PrintError(ERROR, "Error al actualizar el DD del archivo")
				return false
			}

		}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return false
	}
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func EjecutarAdd(path string, name string, sizePartition int, unit string, comando CONSTCOMANDO) {
	tamanio := TamanioTotal(comando, sizePartition, unit)
	if tamanio != -1 {
		if addParticion(path, name, tamanio, comando) {
			PrintAviso(comando, "Se modifico el tamanio de la particion exitosamente")
			return
		} else {
			PrintError(ERROR, "no se pudo modificar el tamanio de la particion")
			return
		}
	} else {
		PrintError(ERROR, "Error al calcular el tamanio finald el size")
		return
	}
}

func addParticion(pathDisco string, namePart string, tamanio int, comando CONSTCOMANDO) bool {

	var mbr *MBRStruct
	mbr = new(MBRStruct) //modificar el tamanio, fin nuevo//
	//EXTRAER MBR//BUSCAR SU PARTICION IGUAL A ESTA	//IGUALARLO A UNA NUEVA PARTITION NUEVA (LIMPIAR)//ORDENAR EL ARREGLO//IR A GUARDAR MBR
	if ExtrarMBR(pathDisco, comando, mbr) {
		PrintAviso(comando, "Se encontro el disco")

		if partitionBool, partition := getParticionByNameDelete(mbr, namePart); partitionBool {

			if changeSize(mbr, partition, tamanio, comando) {

				OrdenarMBRParticiones(mbr)

				if GuardarMBR(comando, mbr, pathDisco) {
					PrintAviso(comando, "Se modifico el tamanio de la particion y el fin exitosamente en la tabla de particiones y se actualizo el disco [Disco: "+pathDisco+", Particion: "+namePart+"]")
					return true
				} else {
					PrintError(ERROR, "Error al guardar el MBR actualizado [Disco: "+pathDisco+", Particion: "+namePart+"]")
					return false
				}

			} else {
				PrintError(ERROR, "No se pudo ejecutar el parametro add correctamente")
				return false
			}

		} else {
			PrintError(ERROR, "Error al extraer una particion del mbr, o posiblemente no existe como [P|E] [Disco: "+pathDisco+", Particion: "+namePart+"]")
			return false
		}

	} else {
		PrintError(ERROR, "Error al extraer el mbr del disco o no existe el disco [Disco: "+pathDisco+", Particion: "+namePart+"]")
		return false
	}
}

//si es negativo ver que quede espacio todavia, si es extendida tomando en cuenta su ebr
//si es positivo ver que todavia existe espacio libre despues de la particion

func changeSize(mbr *MBRStruct, partition *PartitionStruct, tamanio int, comando CONSTCOMANDO) bool {

	if tamanio > 0 { //agregar

		if isFree, valFree := spaceFree(mbr, partition); isFree {

			if tamanio <= valFree.Lsize {

				partition.PartSize = partition.PartSize + int64(tamanio)
				PrintAviso(comando, "Se agrego espacio a la particion exitosamente [Se agrego: "+strconv.Itoa(tamanio)+"]")
				return true

			} else {
				PrintError(ERROR, "el tamanio que deseas agreagar es mayor al espacio libre disponible, solo se tiene disponible [Free size:"+strconv.Itoa(valFree.Lsize)+"]")
				return false
			}

		} else {
			PrintError(ERROR, "No se encontro espacio libre despues de la particion")
			return false
		}

	} else { //disminuir

		//ver que si sea negativo
		if partition.PartType == 'P' {
			if (partition.PartSize + int64(tamanio)) > 0 {

				partition.PartSize = partition.PartSize + int64(tamanio)
				PrintAviso(comando, "Se redujo el espacio de la particion exitosamente nuevo tamanio ["+strconv.Itoa(int(partition.PartSize))+"]")
				return true

			} else {
				PrintError(ERROR, "La particion se quedaria sin espacio, estas reduciendo todo el tamanio y posiblemente mas")
				return false
			}

		} else {
			if (partition.PartSize + int64(tamanio)) > int64(binary.Size(EBRStruct{})) {

				partition.PartSize = partition.PartSize + int64(tamanio)
				PrintAviso(comando, "Se redujo el espacio de la particion exitosamente nuevo tamanio ["+strconv.Itoa(int(partition.PartSize))+"]")
				return true

			} else {
				PrintError(ERROR, "La particion es de tipo extendida y por lo menos debes dejar el espacio sufiente para su ebr inicial")
				return false
			}
		}

	}
}

func spaceFree(mbr *MBRStruct, partition *PartitionStruct) (bool, Libre) {

	start := int64(binary.Size(mbr)) //posArchivo
	end := int64(0)

	ultimoEnd := int64(0)
	listaLibres := make([]Libre, 0)

	for i := range mbr.Partition {

		if mbr.Partition[i].PartStart != -1 { //.PartStatus == 49 (1)(si ocupado) || == 48 (0)(no ocupado)

			end = mbr.Partition[i].PartStart
			if (end - start) > 0 { //hay espacio libre
				libre := Libre{
					Lstart: int(start),
					Lend:   int(end),
					Lsize:  int(end) - int(start),
				}
				listaLibres = append(listaLibres, libre)
			}
			start = mbr.Partition[i].PartStart + mbr.Partition[i].PartSize

			ultimoEnd = mbr.Partition[i].PartStart + mbr.Partition[i].PartSize

		}

	}

	if ultimoEnd < mbr.MbrTamanio {
		libre := Libre{
			Lstart: int(ultimoEnd),
			Lend:   int(mbr.MbrTamanio),
			Lsize:  int(mbr.MbrTamanio) - int(ultimoEnd),
		}
		listaLibres = append(listaLibres, libre)
	}

	for _, libre := range listaLibres {
		if int(partition.PartStart+partition.PartSize) == libre.Lstart {
			return true, libre
		}
	}

	return false, Libre{}
}

func ComandoFind(nodoDisco *NodoDisco, nodoParticion *NodoParticion, mapa map[string]string, comando CONSTCOMANDO) bool {
	arrDir := ArrDir(mapa["PATH"])
	if arrDir != nil { //archivo afuera de la raiz

		if val, found := mapa["NAME"]; found {

			if strings.Contains(val, ".") { //buscar el archivo
				findArchivo(nodoDisco, nodoParticion, comando, arrDir)
				return true
			} else { //es para carpeta
				findDir(nodoDisco, nodoParticion, comando, arrDir)
				return true
			}

		} else {
			PrintError(ERROR, "El parametro name no se encuentra dentro de la sentencia")
			return false
		}
	} else {
		PrintError(ERROR, "Error con el parametro path")
		return false
	}
}

func findArchivo(nodoDisco *NodoDisco, nodoParticion *NodoParticion, comando CONSTCOMANDO, arrDir [][50]byte) {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		contadorArr := int64(0)
		posAvdBitmapRoot := int64(0)
		_, _, _, valPos := MkdirRecorrer(nodoDisco.path, avd, arrDir, contadorArr, sb.SbApArbolDirectorio, posAvdBitmapRoot, sb.SbSizeStructArbolDirectorio, 0, 1, comando)

		if valPos != -1 {

			_, texto := findCompleteFromArchivo(comando, nodoDisco.path, avd, sb.SbApArbolDirectorio, valPos, sb.SbSizeStructArbolDirectorio, sb, "_")
			PrintAviso(comando, "Resultado del comando Fin: ")
			fmt.Println(texto)
			return

		} else {
			PrintError(ERROR, "Error al encontrar la posicion de la carpeta deseada")
			return
		}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return
	}
}

func findCompleteFromArchivo(comando CONSTCOMANDO, path string, avd *AVDStruct, startAVDStructs int64, posAVDDisco int64, sizeAVDStruct int64, sb *SuperBootStruct, tabulador string) (bool, string) {

	avd = new(AVDStruct)
	if ExtrarAVD(path, comando, avd, int(startAVDStructs+(posAVDDisco*sizeAVDStruct))) {

		tabulador = tabulador + "_"

		//texto := string(bytes.Trim(avd.AVDNombreDirectorio[:], "0")) + "\n"
		texto := ""
		textoextra := ""

		for _, val := range avd.AVDApArraySubdirectorios { //mas carpetas
			if val != -1 { //==#

				if valBool1, valTexto1 := findCompleteFromArchivo(comando, path, avd, startAVDStructs, val, sizeAVDStruct, sb, tabulador); valBool1 {
					textoextra += valTexto1
				}
			}
		}
		if avd.AVDApArbolVirtualDirectorio != -1 { //tiene un indirecto
			tabulador = "_" //reinicio
			if valBool2, valTexto2 := findCompleteFromArchivo(comando, path, avd, startAVDStructs, avd.AVDApArbolVirtualDirectorio, sizeAVDStruct, sb, tabulador); valBool2 {
				texto += valTexto2
			}
		}

		if avd.AVDApDetalleDirectorio != -1 { //si tiene un DD
			var dd *DDStruct
			dd = new(DDStruct)
			if valBool3, valTexto3 := findArchivoDD(path, sb, avd.AVDApDetalleDirectorio, dd, comando, tabulador+"_"); valBool3 {
				texto += valTexto3
			}
		}

		return true, texto + textoextra

	} else {
		PrintError(ERROR, "Error al extraer el [AVD: Raiz]")
		return false, "" //una estructura vacia
	}
}

//RepTreeCompleteDetalleDtexto retorna el texto para un dd y toda su info
func findArchivoDD(path string, sb *SuperBootStruct, posBitmapDD int64, dd *DDStruct, comando CONSTCOMANDO, tabulador string) (bool, string) {

	dd = new(DDStruct)
	if ExtrarDD(path, comando, dd, int(sb.SbApDetalleDirectorio+(posBitmapDD*sb.SbSizeStructDetalleDirectorio))) {

		texto := ""

		for _, val := range dd.DDArrayFile {
			if string(bytes.Trim(val.DDFileNombre[:], "0")) != "" { //tiene nombre
				texto += tabulador + "|" + string(bytes.Trim(val.DDFileNombre[:], "0")) + "\n"
			}
		}

		if dd.DDApDetalleDirectorio != -1 { //si tiene un indirecto
			if valBool1, valTexto1 := findDD(path, sb, dd.DDApDetalleDirectorio, dd, comando, tabulador); valBool1 {
				texto += valTexto1
			}
		}

		return true, texto

	} else {
		PrintError(ERROR, "Error al extraer el dd del archivo para la carpeta")
		return false, ""
	}
}

//RepTreeComplete reporte completo del sistema de archivos
func findDir(nodoDisco *NodoDisco, nodoParticion *NodoParticion, comando CONSTCOMANDO, arrDir [][50]byte) {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		contadorArr := int64(0)
		posAvdBitmapRoot := int64(0)
		_, _, _, valPos := MkdirRecorrer(nodoDisco.path, avd, arrDir, contadorArr, sb.SbApArbolDirectorio, posAvdBitmapRoot, sb.SbSizeStructArbolDirectorio, 0, 1, comando)

		if valPos != -1 {

			_, texto := findCompleteFrom(comando, nodoDisco.path, avd, sb.SbApArbolDirectorio, valPos, sb.SbSizeStructArbolDirectorio, sb, "_")
			PrintAviso(comando, "Resultado del comando Fin: ")
			fmt.Println(texto)
			return

		} else {
			PrintError(ERROR, "Error al encontrar la posicion de la carpeta deseada")
			return
		}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return
	}
}

func findCompleteFrom(comando CONSTCOMANDO, path string, avd *AVDStruct, startAVDStructs int64, posAVDDisco int64, sizeAVDStruct int64, sb *SuperBootStruct, tabulador string) (bool, string) {

	avd = new(AVDStruct)
	if ExtrarAVD(path, comando, avd, int(startAVDStructs+(posAVDDisco*sizeAVDStruct))) {

		tabulador = tabulador + "_"

		texto := string(bytes.Trim(avd.AVDNombreDirectorio[:], "0")) + "\n"
		textoextra := ""

		for _, val := range avd.AVDApArraySubdirectorios { //mas carpetas
			if val != -1 { //==#

				if valBool1, valTexto1 := findCompleteFrom(comando, path, avd, startAVDStructs, val, sizeAVDStruct, sb, tabulador); valBool1 {
					textoextra += tabulador + "|" + valTexto1
				}
			}
		}
		if avd.AVDApArbolVirtualDirectorio != -1 { //tiene un indirecto
			tabulador = "_" //reinicio
			if valBool2, valTexto2 := findCompleteFrom(comando, path, avd, startAVDStructs, avd.AVDApArbolVirtualDirectorio, sizeAVDStruct, sb, tabulador); valBool2 {
				texto += valTexto2
			}
		}

		if avd.AVDApDetalleDirectorio != -1 { //si tiene un DD
			var dd *DDStruct
			dd = new(DDStruct)
			if valBool3, valTexto3 := findDD(path, sb, avd.AVDApDetalleDirectorio, dd, comando, tabulador+"_"); valBool3 {
				texto += valTexto3
			}
		}

		return true, texto + textoextra

	} else {
		PrintError(ERROR, "Error al extraer el [AVD: Raiz]")
		return false, "" //una estructura vacia
	}
}

//RepTreeCompleteDetalleDtexto retorna el texto para un dd y toda su info
func findDD(path string, sb *SuperBootStruct, posBitmapDD int64, dd *DDStruct, comando CONSTCOMANDO, tabulador string) (bool, string) {

	dd = new(DDStruct)
	if ExtrarDD(path, comando, dd, int(sb.SbApDetalleDirectorio+(posBitmapDD*sb.SbSizeStructDetalleDirectorio))) {

		texto := ""

		for _, val := range dd.DDArrayFile {
			if string(bytes.Trim(val.DDFileNombre[:], "0")) != "" { //tiene nombre
				texto += tabulador + "|" + string(bytes.Trim(val.DDFileNombre[:], "0")) + "\n"
			}
		}

		if dd.DDApDetalleDirectorio != -1 { //si tiene un indirecto
			if valBool1, valTexto1 := findDD(path, sb, dd.DDApDetalleDirectorio, dd, comando, tabulador); valBool1 {
				texto += valTexto1
			}
		}

		return true, texto

	} else {
		PrintError(ERROR, "Error al extraer el dd del archivo para la carpeta")
		return false, ""
	}
}

func ComandoRen(nodoDisco *NodoDisco, nodoParticion *NodoParticion, rutaFile string, newNombre string, comando CONSTCOMANDO) {
	rutaFile = strings.ReplaceAll(rutaFile, "\"", "")
	arrDir := ArrDir(rutaFile)
	archivo := ""
	if arrDir != nil {

		isArchivo := false
		if strings.Contains(rutaFile, ".") { //ruta de archivo

			archivo = string(arrDir[len(arrDir)-1][:])
			arrDir = arrDir[:len(arrDir)-1]
			isArchivo = true

		} else { //ruta de carpeta

			arrDir = arrDir[:]
			isArchivo = false
		}

		if ejecturarRen(nodoDisco, nodoParticion, arrDir, archivo, newNombre, comando, isArchivo) {

			PrintAviso(comando, "Se modifico el nombre exitosamente")
			return

		} else {
			PrintError(ERROR, "error al modificar el nombre")
			return
		}
	} else {
		PrintError(ERROR, "Existe algun error con el parametro ruta")
		return
	}
}

func ejecturarRen(nodoDisco *NodoDisco, nodoParticion *NodoParticion, arrDir [][50]byte, archivo string, newNombre string, comando CONSTCOMANDO, isArchivo bool) bool {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		listaAVDS := make([]*avdTree, 0) //lista de avdTress
		treeTres(nodoDisco.path, sb, avd, arrDir, archivo, 0, 0, comando, &listaAVDS)

		if len(listaAVDS) == 0 {
			PrintAviso(comando, "la ruta no existe o no existe completa")
			return false
		}

		if !isArchivo { //carpeta

			for _, avd := range listaAVDS {
				if avd.avdUnidad.AVDNombreDirectorio == arrDir[len(arrDir)-1] {
					avd.avdUnidad.AVDNombreDirectorio = [50]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
					copy(avd.avdUnidad.AVDNombreDirectorio[:], newNombre)
					GuardarAVD(comando, avd.avdUnidad, nodoDisco.path, int(sb.SbApArbolDirectorio+(avd.posBitmapAVD*sb.SbSizeStructArbolDirectorio)))
				}
			}

			PrintAviso(comando, "Se termino el proceso de modificacion de nombre de carpeta")
			return true

		} else { //archivo

			for _, avd := range listaAVDS {
				for _, dd := range *avd.ListaDD {
					for _, inodo := range *dd.ListaInodo {
						for i := range dd.ddUnidad.DDArrayFile { //desuniendo
							if dd.ddUnidad.DDArrayFile[i].DDFileApInodo == inodo.posBitmapInodo { //en la pos del arreglo
								dd.ddUnidad.DDArrayFile[i].DDFileNombre = [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
								copy(dd.ddUnidad.DDArrayFile[i].DDFileNombre[:], newNombre)
								GuardarDD(comando, dd.ddUnidad, nodoDisco.path, int(sb.SbApDetalleDirectorio+(dd.posBitmapDD*sb.SbSizeStructDetalleDirectorio)))
							}
						}
					}
				}
			}

			PrintAviso(comando, "Se termino el proceso de modificacion de nombre de archivo")
			return true

		}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return false
	}
}

func ComandoRm(nodoDisco *NodoDisco, nodoParticion *NodoParticion, rutaFile string, comando CONSTCOMANDO) {
	rutaFile = strings.ReplaceAll(rutaFile, "\"", "")
	arrDir := ArrDir(rutaFile)
	archivo := ""
	if arrDir != nil {

		isArchivo := false
		if strings.Contains(rutaFile, ".") { //ruta de archivo

			archivo = string(arrDir[len(arrDir)-1][:])
			arrDir = arrDir[:len(arrDir)-1]
			isArchivo = true

		} else { //ruta de carpeta

			arrDir = arrDir[:]
			isArchivo = false
		}

		if ejecutarRM(nodoDisco, nodoParticion, arrDir, archivo, comando, isArchivo) {

			PrintAviso(comando, "Se elimino la ruta exitosamente")
			return

		} else {
			PrintError(ERROR, "error al eliminar la ruta")
			return
		}
	} else {
		PrintError(ERROR, "Existe algun error con el parametro ruta")
		return
	}
}

//[bd de archivo]
//[archivo desunir de dd]//si era el ultimo
//[dd de avd]
func ejecutarRM(nodoDisco *NodoDisco, nodoParticion *NodoParticion, arrDir [][50]byte, archivo string, comando CONSTCOMANDO, isArchivo bool) bool {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		listaAVDS := make([]*avdTree, 0) //lista de avdTress
		treeTres(nodoDisco.path, sb, avd, arrDir, archivo, 0, 0, comando, &listaAVDS)

		if len(listaAVDS) == 0 {
			PrintAviso(comando, "la ruta no existe o no existe completa")
			return false
		}

		if !isArchivo { //carpeta

			posToClean := int64(-1)
			for _, avd := range listaAVDS { //como vengo de 0 ... +, seria hasta llegar a esta carpeta
				if avd.avdUnidad.AVDNombreDirectorio == arrDir[len(arrDir)-1] { //la carpeta que busco de aqui en adelante

					//avd que estoy eliminando
					posToClean = avd.posBitmapAVD
					//limpiar las carpetas que tenia
					for i := range avd.avdUnidad.AVDApArraySubdirectorios { //cada aptr ir a limpiarlo al bitmap
						//TODO:IR A LIMPIAR LAS CARPETAS QUE CONTIENEN ESTAS QUE ESTOY ELIMINANDO...RECURSIVO
						if avd.avdUnidad.AVDApArraySubdirectorios[i] != -1 {
							limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitMapArbolDirectorio, sb.SbArbolVirtualCount, "AVD", avd.avdUnidad.AVDApArraySubdirectorios[i], comando)
						}
					}
					//limpiar el indirecto la carpeta subdirectorio
					if avd.avdUnidad.AVDApArbolVirtualDirectorio != -1 {
						limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitMapArbolDirectorio, sb.SbArbolVirtualCount, "AVD", avd.avdUnidad.AVDApArbolVirtualDirectorio, comando)
					}

					//desunir limpiar su anterior (en este caso la sig pos)...afuera

					//cuando ya puedo limpiar su posicion en la tabla
					limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitMapArbolDirectorio, sb.SbArbolVirtualCount, "AVD", avd.posBitmapAVD, comando)

					for _, dd := range *avd.ListaDD {
						limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapDetalleDirectorio, sb.SbDetalleDirectorioCount, "DD", dd.posBitmapDD, comando)
						for _, inodo := range *dd.ListaInodo {
							limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapTablaInodo, sb.SbInodosCount, "INODO", inodo.posBitmapInodo, comando)
							for _, bd := range *inodo.ListaBD {
								limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapBloques, sb.SbBloquesCount, "BD", bd.posBitmapBD, comando)
							}
						}
					}
				}
			}

			//desunir
			for _, avd := range listaAVDS {
				for i := range avd.avdUnidad.AVDApArraySubdirectorios {
					if posToClean != -1 {
						if avd.avdUnidad.AVDApArraySubdirectorios[i] == posToClean {
							avd.avdUnidad.AVDApArraySubdirectorios[i] = -1
							GuardarAVD(comando, avd.avdUnidad, nodoDisco.path, int(sb.SbApArbolDirectorio+(avd.posBitmapAVD*sb.SbSizeStructArbolDirectorio)))
							//var avdfff *AVDStruct
							//avdfff = new(AVDStruct)
							//ExtrarAVD(nodoDisco.path, comando, avdfff, int(sb.SbApArbolDirectorio + (avd.posBitmapAVD * sb.SbSizeStructArbolDirectorio)))
						}
					}
				}
			}

			PrintAviso(comando, "Se termino el proceso de eliminacion para una carpeta y todo su contenido")
			return true

		} else { //archivo

			for _, avd := range listaAVDS {
				//limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitMapArbolDirectorio, sb.SbArbolVirtualCount, "AVD", avd.posBitmapAVD, comando)
				for _, dd := range *avd.ListaDD {
					//limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapDetalleDirectorio, sb.SbDetalleDirectorioCount, "DD", dd.posBitmapDD, comando)
					for _, inodo := range *dd.ListaInodo {

						for i := range dd.ddUnidad.DDArrayFile { //desuniendo
							if dd.ddUnidad.DDArrayFile[i].DDFileApInodo == inodo.posBitmapInodo { //en la pos del arreglo, elimino el/los inodos que tengo en mi arreglo
								dd.ddUnidad.DDArrayFile[i].DDFileNombre = [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
								dd.ddUnidad.DDArrayFile[i].DDFileDateModificacion = [25]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
								dd.ddUnidad.DDArrayFile[i].DDFileDateCreacion = [25]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
								dd.ddUnidad.DDArrayFile[i].DDFileApInodo = -1
								GuardarDD(comando, dd.ddUnidad, nodoDisco.path, int(sb.SbApDetalleDirectorio+(dd.posBitmapDD*sb.SbSizeStructDetalleDirectorio)))
							}
						}

						limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapTablaInodo, sb.SbInodosCount, "INODO", inodo.posBitmapInodo, comando)

						for _, bd := range *inodo.ListaBD { //eliminando los bd
							limpiarBitmapyEscribirlo(nodoDisco.path, sb.SbApBitmapBloques, sb.SbBloquesCount, "BD", bd.posBitmapBD, comando)
						}
					}
				}
			}

			PrintAviso(comando, "Se termino el proceso de eliminacion de archivo")
			return true

		}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return false
	}
}

func limpiarBitmapyEscribirlo(pathDisco string, sbApBitmap int64, tamanioCoun int64, nameBitmap string, posBitmap int64, comando CONSTCOMANDO) bool {
	bitmapBool, bitmapVal := ExtrarBitmap(pathDisco, comando, int(sbApBitmap), int(tamanioCoun), nameBitmap)

	if bitmapBool {

		if int(posBitmap) < len(bitmapVal) {
			bitmapVal[posBitmap] = 0
		}
		if GuardarBitmap(comando, bitmapVal, pathDisco, int(sbApBitmap), "") {

			PrintAviso(comando, "Bitmap actualizado correctamente [Bitmap: "+nameBitmap+"]")
			return true

		} else {
			PrintError(ERROR, "Error al momento de ocupar los espacios en el bitmap y grabarlos en la particion del disco")
			return false
		}

	} else {
		PrintError(ERROR, "Error al extraer el bitmap correspondiente [Bitmap: "+nameBitmap+"]")
		return false
	}
}

//desmontandoParticion ejecuta el comando Unmount para un id
//func desmontandoParticion(mountList *[]*NodoDisco, letraDisco string, id string, comando CONSTCOMANDO) {

func ComandoCat(nodoDis *NodoDisco, nodoPart *NodoParticion, mapa map[string]string, comando CONSTCOMANDO) {
	textoFinal := ""
	for k, ruta := range mapa {
		if k != "INSTRUCCION" && k != "CAT" && k != "ID" {

			if strings.HasPrefix(k, "FILE") {

				if isok, texto := tree(nodoDis, nodoPart, ruta, comando); isok {
					textoFinal += texto

				} else {
					PrintError(ERROR, "Existio un error para generar el archivo del reporte cat")
					return
				}

			} else {
				PrintError(comando, "Este file no tiene la sintaxis correcta [file: "+k+"]")
				return
			}
		}
	}

	PrintAviso(comando, "Exitosamente se ejecuto el comando cat")

	if generadorImagen("/home/user/reports/cat.dot", textoFinal, comando) {
		PrintAviso(comando, "Imagen generada correctamente para el archivo [Nombre:cat]")
		return
	} else {
		PrintError(ERROR, "Error al generar el reporte del comando cat [Nombre: cat]")
		return
	}

}

func tree(nodoDisco *NodoDisco, nodoParticion *NodoParticion, rutaFile string, comando CONSTCOMANDO) (bool, string) {

	rutaFile = strings.ReplaceAll(rutaFile, "\"", "")
	arrDir := ArrDir(rutaFile)
	if arrDir != nil {
		archivo := string(arrDir[len(arrDir)-1][:])
		arrDir = arrDir[:len(arrDir)-1]
		return treeDos(nodoDisco, nodoParticion, arrDir, archivo, comando)
	} else {
		PrintError(ERROR, "Existe algun error con el parametro ruta")
		return false, ""
	}
}

func treeDos(nodoDisco *NodoDisco, nodoParticion *NodoParticion, arrDir [][50]byte, archivo string, comando CONSTCOMANDO) (bool, string) {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		listaAVDS := make([]*avdTree, 0) //lista de avdTress
		treeTres(nodoDisco.path, sb, avd, arrDir, archivo, 0, 0, comando, &listaAVDS)

		if len(listaAVDS) == 0 {
			PrintAviso(comando, "la ruta no existe o no existe completa")
		}

		texto := ""
		for _, avd := range listaAVDS {
			for _, dd := range *avd.ListaDD {
				for i := len(*dd.ListaInodo) - 1; i >= 0; i-- {
					var inodo *inodoTree
					inodo = new(inodoTree)
					inodo = (*dd.ListaInodo)[i]
					for _, bd := range *inodo.ListaBD {
						PrintAviso(comando, string(bd.bdUnidad.DbData[:]))
						texto += string(bd.bdUnidad.DbData[:]) + "\n"
					}
				}
			}
		}
		return true, texto
		//for _, inodo := range *dd.ListaInodo {
		//	for _, bd := range *inodo.ListaBD {
		//		fmt.Println(comando, string(bd.bdUnidad.DbData[:]))
		//		texto += string(bd.bdUnidad.DbData[:]) + "\n"
		//	}
		//}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return false, ""
	}
}

func treeTres(pathDisco string, sb *SuperBootStruct, avd *AVDStruct, arrDir [][50]byte, archivo string, contadorArr int64, posBitmapAVD int64, comando CONSTCOMANDO, listaAVDS *[]*avdTree) (bool, string) {

	avd = new(AVDStruct)
	if ExtrarAVD(pathDisco, comando, avd, int(sb.SbApArbolDirectorio+(posBitmapAVD*sb.SbSizeStructArbolDirectorio))) {

		if int(contadorArr) < len(arrDir) {

			if avd.AVDNombreDirectorio == arrDir[contadorArr] {
				//ADENTRO DE EL AVD
				contadorArr++

				var avdArbol *avdTree //una unidad para la lista de avd
				avdArbol = new(avdTree)
				avdArbol.avdUnidad = avd //su valor
				avdArbol.posBitmapAVD = posBitmapAVD

				avdArbol.ListaDD = new([]*ddTree)      //inicializo la lista de DD
				*avdArbol.ListaDD = make([]*ddTree, 0) //su lista

				if int(contadorArr) != len(arrDir) { //sino

					for _, val := range avd.AVDApArraySubdirectorios {
						if val != -1 {
							if valBool, _ := treeTres(pathDisco, sb, avd, arrDir, archivo, contadorArr, val, comando, listaAVDS); valBool {

							}
						}
					}
					if avd.AVDApArbolVirtualDirectorio != -1 { //si tiene #
						contadorArr--
						if valBool, _ := treeTres(pathDisco, sb, avd, arrDir, archivo, contadorArr, avd.AVDApArbolVirtualDirectorio, comando, listaAVDS); valBool {

						}
					}

				} else { //==len()//seguir por explorar su dd
					PrintAviso(comando, "Ya fue encontrado el PATH completo")
					if avd.AVDApDetalleDirectorio != -1 {

						var dd *DDStruct
						dd = new(DDStruct)

						nombreComodin := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
						copy(nombreComodin[:], archivo)

						if valBool3, _ := treeCuatro(pathDisco, sb, avd.AVDApDetalleDirectorio, dd, comando, nombreComodin, avdArbol.ListaDD); valBool3 {

						}
					}
				}

				*listaAVDS = append(*listaAVDS, avdArbol)
				return true, ""
				//ADENTRO DE EL AVD
			} else {
				return false, "" //false
			}

		} else {
			return false, "" //false
		}

	} else {
		PrintError(ERROR, "Error al extraer el [AVD]")
		return false, ""
	}
}

func treeCuatro(path string, sb *SuperBootStruct, posBitmapDD int64, dd *DDStruct, comando CONSTCOMANDO, DDFileNombre [16]byte, ListaDD *[]*ddTree) (bool, string) {

	dd = new(DDStruct)
	if ExtrarDD(path, comando, dd, int(sb.SbApDetalleDirectorio+(posBitmapDD*sb.SbSizeStructDetalleDirectorio))) {

		var ddArbol *ddTree //inicializo la unidad
		ddArbol = new(ddTree)
		ddArbol.ddUnidad = dd //su valor
		ddArbol.posBitmapDD = posBitmapDD

		ddArbol.ListaInodo = new([]*inodoTree)      //inicializo su la lista de Inodos
		*ddArbol.ListaInodo = make([]*inodoTree, 0) //su lista

		bandera := false
		for _, val := range dd.DDArrayFile { //ir a traer sus inodos
			if (val.DDFileApInodo != -1) && (val.DDFileNombre == DDFileNombre) { //==este ocupado y sera ==nombre
				bandera = true
				var inodo *InodoStruct
				inodo = new(InodoStruct)

				if valBool1, _ := treeCinco(path, comando, inodo, val.DDFileApInodo, sb, ddArbol.ListaInodo); valBool1 {

				}
			}
		}

		if dd.DDApDetalleDirectorio != -1 { //si tiene un indirecto
			if valBool1, _ := treeCuatro(path, sb, dd.DDApDetalleDirectorio, dd, comando, DDFileNombre, ListaDD); valBool1 {

			}
		}

		if bandera {
			*ListaDD = append(*ListaDD, ddArbol)
			return true, ""
		} else {
			PrintAviso(comando, "No existe un archivo con ese nombre amigo")
			return false, ""
		}

	} else {
		PrintError(ERROR, "Error al extraer el dd del archivo para la carpeta")
		return false, ""
	}
}

//RepInodoTexto texto para un inodo
func treeCinco(path string, comando CONSTCOMANDO, inodo *InodoStruct, posBitmapInodo int64, sb *SuperBootStruct, ListaInodo *[]*inodoTree) (bool, string) {
	inodo = new(InodoStruct)
	if ExtrarInodo(path, comando, inodo, int(sb.SbApTablaInodo+(posBitmapInodo*sb.SbSizeStructInodo))) {

		var inodoArbol *inodoTree //inicializo la unidad
		inodoArbol = new(inodoTree)
		inodoArbol.inodoUnidad = inodo //su valor
		inodoArbol.posBitmapInodo = posBitmapInodo

		inodoArbol.ListaBD = new([]*bdTree)      //inicializo su la lista de Inodos
		*inodoArbol.ListaBD = make([]*bdTree, 0) //su lista

		for _, val := range inodo.IArrayBloques { //con sus bloques
			if val != -1 {
				var bd *BloqueDeDatosStruct

				if valBool1, _ := treeSeis(path, comando, bd, val, sb, inodoArbol.ListaBD); valBool1 {

				}
			}
		}

		if inodo.IApIndirecto != -1 { //si tiene un indirecto
			if valBool1, _ := treeCinco(path, comando, inodo, inodo.IApIndirecto, sb, ListaInodo); valBool1 {

			}
		}

		*ListaInodo = append(*ListaInodo, inodoArbol)
		return true, ""

	} else {
		PrintError(ERROR, "Error al extraer el inodo del archivo")
		return false, ""
	}
}

//RepBDTexto extrae el texto para un BD
func treeSeis(path string, comando CONSTCOMANDO, bd *BloqueDeDatosStruct, posBitmapBD int64, sb *SuperBootStruct, ListaBD *[]*bdTree) (bool, string) {
	bd = new(BloqueDeDatosStruct)
	if ExtrarBD(path, comando, bd, int(sb.SbApBloques+(posBitmapBD*sb.SbSizeStructBloque))) {

		var bdArbol *bdTree //inicializo la unidad
		bdArbol = new(bdTree)
		bdArbol.posBitmapBD = posBitmapBD

		bdArbol.bdUnidad = bd
		*ListaBD = append(*ListaBD, bdArbol)

		return true, ""

	} else {
		PrintError(ERROR, "Error al extraer un BD del archivo")
		return false, ""
	}
}

//_________________________________________________

//EjecutarDelete borra una particion
func EjecutarDelete(pathDisco string, namePart string, typeDelete string, comando CONSTCOMANDO) {
	if formatearParticion(pathDisco, namePart, typeDelete, comando) {
		PrintAviso(comando, "Se formateo exitosamente la particion")
		return
	} else {
		PrintError(ERROR, "no se pudo formatear la particion")
		return
	}
}

//formatearParticion formatea una particion primaria, extendida
func formatearParticion(pathDisco string, namePart string, tipoFormateo string, comando CONSTCOMANDO) bool {

	var mbr *MBRStruct
	mbr = new(MBRStruct)
	//EXTRAER MBR//BUSCAR SU PARTICION IGUAL A ESTA	//IGUALARLO A UNA NUEVA PARTITION NUEVA (LIMPIAR)//ORDENAR EL ARREGLO//IR A GUARDAR MBR
	if ExtrarMBR(pathDisco, comando, mbr) {
		PrintAviso(comando, "Se encontro el disco ")

		if partitionBool, partition := getParticionByNameDelete(mbr, namePart); partitionBool {
			sizePart := partition.PartSize
			startPart := partition.PartStart
			construirPartition(partition, '0', '0', '0', -1, 0, "")
			PrintAviso(comando, "Se le limpio la particion en la tabla de particiones")

			OrdenarMBRParticiones(mbr)

			if GuardarMBR(comando, mbr, pathDisco) {
				PrintAviso(comando, "Se formato exitosamente la particion en la tabla de particiones y se actualizo el disco [Disco: "+pathDisco+", Particion: "+namePart+"]")
				if tipoFormateo == "FAST" {
					PrintAviso(comando, "El formateo Fast fue ejecutado exitosamente")
					return true
				} else { //==FULL
					if limpiarArchivo(comando, pathDisco, int(sizePart), int(startPart)) {
						PrintAviso(comando, "El formateo Full fue ejecutado exitosamente")
						return true
					} else {
						PrintError(ERROR, "Error al formatear el area de la particion en el disco para completar el formateo full")
						return false
					}
				}
			} else {
				PrintError(ERROR, "Error al guardar el MBR actualizado [Disco: "+pathDisco+", Particion: "+namePart+"]")
				return false
			}

		} else {
			PrintError(ERROR, "Error al extraer una particion del mbr, o posiblemente no existe [Disco: "+pathDisco+", Particion: "+namePart+"]")
			return false
		}

	} else {
		PrintError(ERROR, "Error al extraer el mbr del disco o no existe el disco [Disco: "+pathDisco+", Particion: "+namePart+"]")
		return false
	}
}

//getParticionByName obtiene una particion [principal] si existe con el mismo nombre
func getParticionByNameDelete(mbr *MBRStruct, name string) (bool, *PartitionStruct) {
	nombre := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	copy(nombre[:], name)
	for i := range mbr.Partition {
		if mbr.Partition[i].PartStatus != '0' {
			if mbr.Partition[i].PartName == nombre { //int = 48//byte = '0'  || int = 0// int = 0
				return true, &mbr.Partition[i] //TODO: VERIFICAR SI MODIFICO AL ORIGINAL
			}
		}
	}
	return false, nil
}

//TODO CONFIRMAR PROCESO DE CREAR UN ARCHIVO DE TAMANIO 0 //size==0 || contenido=="", pero siempre iguales y el size manda
//TODO LA POSICIONBITMAP DEL DD SINO LO USO LO TENGO QUE IR A DEJAR LIBRE
//TODO VER QUE PASA CUANDO EL ARREGLO ES DE TAMA = 0 Y RETORNA UN NIL, MAS QUE TODO EN EL CASO DE UN BD

//comandoMkfile comando mkfile
func comandoMkfile(pathDisco string, nodoPart *NodoParticion, sb *SuperBootStruct, sbCopia *SuperBootStruct, arrArchivo [][50]byte, rutaMkdir string, resultisP bool, size int, contenido string, comando CONSTCOMANDO, nameProper string, partFit byte) bool {

	necesitoBD := math.Ceil(float64(size) / 25) //aprox al mayor
	necesitoInodo := math.Ceil(necesitoBD / 4)  //aprox al mayor
	if necesitoInodo == 0 {
		necesitoInodo = 1
	}
	necesitoDD := 1 //si no existe si ==1|| si ya existe-->||no lleno, uso el mismo ==0|| si lleno, ==1||

	archivo := string(arrArchivo[len(arrArchivo)-1][:])
	arrDir := arrArchivo[:len(arrArchivo)-1]

	if strings.Contains(archivo, ".") {

		bdBool, arrPosicionesBD := obtenerPosAjusteMkfile(int(necesitoBD), "BD", int(sb.SbApBitmapBloques), int(sb.SbBloquesCount), comando, pathDisco, nameProper, partFit)
		contadorPosicionesBD := 0 //POSICIONES BD

		inodoBool, arrPosicionesInodo := obtenerPosAjusteMkfile(int(necesitoInodo), "INODO", int(sb.SbApBitmapTablaInodo), int(sb.SbInodosCount), comando, pathDisco, nameProper, partFit)
		contadorPosicionesInodo := 0 //POSICIONES INODO

		ddBool, arrPosicionesDD := obtenerPosAjusteMkfile(necesitoDD, "DD", int(sb.SbApBitmapDetalleDirectorio), int(sb.SbDetalleDirectorioCount), comando, pathDisco, nameProper, partFit)
		contadorPosicionesDD := 0 //POSICIONES DD

		arrPosicionesAVD := obtenerPosAjuste(pathDisco, sb, rutaMkdir, nameProper, comando, partFit)
		contadorPosicionesAVD := 0 //POSICIONES AVD

		if arrPosicionesAVD != nil {
			if bdBool && inodoBool && ddBool {
				//-----------------carpetas
				var avdNew *AVDStruct
				avdNew = new(AVDStruct)
				contadorArr := int64(0)
				posAvdBitmapRoot := int64(0)
				countFree := 1
				firstFree := 0

				if mkdirCrear(pathDisco, avdNew, arrDir, contadorArr, sb.SbApArbolDirectorio, posAvdBitmapRoot, sb.SbSizeStructArbolDirectorio, countFree, firstFree, comando, resultisP, arrPosicionesAVD, contadorPosicionesAVD, nameProper) {
					PrintAviso(comando, "EXCELENTE EDUARDO acabas de crear una ruta de carpetas")

					_, _, _, valPos := MkdirRecorrer(pathDisco, avdNew, arrDir, contadorArr, sb.SbApArbolDirectorio, posAvdBitmapRoot, sb.SbSizeStructArbolDirectorio, 0, 1, comando)

					if valPos != -1 {

						if ExtrarAVD(pathDisco, comando, avdNew, int(sb.SbApArbolDirectorio+(valPos*sb.SbSizeStructArbolDirectorio))) {

							//aqui ya tengo la carpeta contenedora, //creo, uno, retorno la direccionDD
							if valBoolCrear, posBitmapDD := DDcrear(pathDisco, valPos, sb, avdNew, arrPosicionesDD, comando); valBoolCrear {
								//-------------------------Archivos
								var ddNew *DDStruct
								ddNew = new(DDStruct) //avdNew/Existente//archivo//size//contenido//structs//arreglos pos//contadores
								if mkfileCrear(pathDisco, sb,
									ddNew,
									posBitmapDD, -1, -1,
									arrPosicionesBD, contadorPosicionesBD,
									arrPosicionesInodo, contadorPosicionesInodo,
									arrPosicionesDD, contadorPosicionesDD,
									archivo, size, contenido, -1, -1, false, comando, nameProper) {

									PrintAviso(comando, "EXCELENTE EDUARDO acabas de crear un archivo")
									return true

								} else {
									PrintError(ERROR, "Existio algun inconveniente para crear un archivo")
									return false
								}
								return true
							} else {
								PrintError(ERROR, "Error al crear el Dd para el AVD")
								return false
							}

						} else {
							PrintError(ERROR, "Error al extraer la carpeta para trabajarla donde debe de ir el archivo")
							return false
						}

					} else {
						PrintError(ERROR, "Error al extraer la posicionBitmap de la carpeta donde debe de ir el archivo")
						return false
					}

				} else {
					PrintError(ERROR, "Existio algun inconveniente al crear una ruta")
					return false
				}

			} else {
				PrintError(ERROR, "Error al calcular las posiciones en el bitmap con el ajuste solicitado [BD: "+strconv.FormatBool(bdBool)+", Inodo: "+strconv.FormatBool(inodoBool)+", DD: "+strconv.FormatBool(ddBool)+"]")
				return false
			}
		} else {
			PrintError(ERROR, "Error al calcular las posiciones en el bitmap con el ajuste solicitado para las carpetas")
			return false
		}

	} else {
		PrintError(ERROR, "el archivo que intentas crear no tiene el formato correcto [Nombre: "+archivo+"]")
		return false
	}
}

//avdNew *AVDStruct, bd *BloqueDeDatosStruct, inodo *InodoStruct,

//mkfileCrear crea un archivo
func mkfileCrear(pathDisco string, sb *SuperBootStruct,
	dd *DDStruct,
	posBitmapDD int64, posBitmapInodo int64, posBitmapBD int64,
	arrPosicionesBD []int, contadorPosicionesBD int,
	arrPosicionesInodo []int, contadorPosicionesInodo int,
	arrPosicionesDD []int, contadorPosicionesDD int,
	archivo string, size int, contenido string,
	encontreFree int64, posFree int64, repetido bool,
	comando CONSTCOMANDO, nameProper string) bool {

	encontreFree = -1
	repetido = false
	posFree = -1
	dd = new(DDStruct)

	if ExtrarDD(pathDisco, comando, dd, int(sb.SbApDetalleDirectorio+(posBitmapDD*sb.SbSizeStructDetalleDirectorio))) {

		for i, val := range dd.DDArrayFile { //todos los aptrs [archivos(inodos)]
			if val.DDFileApInodo == -1 {

				if encontreFree == -1 {
					posFree = int64(i)
					encontreFree++
				}

			} else { //hay un archivo aqui...
				nombreByteArray := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
				copy(nombreByteArray[:], archivo)
				if val.DDFileNombre == nombreByteArray { //ya existe
					repetido = true
					posBitmapInodo = val.DDFileApInodo //enlace al archivo
				}
			}
		}

		if repetido {
			//pathDisco string, inodo *InodoStruct, sb *SuperBootStruct,
			//	posBitmapInodo int64, //pos courier
			//	arrPosicionesBD []int, contadorPosicionesBD int64, //crear nuevos BD
			//	arrPosicionesInodo []int, contadorPosicionesInodo int64, //crear Indirectos
			//	size int64, contenido string, repetido bool, comando CONSTCOMANDO, nameProper string, posOcupados []int
			posOcupados := make([]int, 0)

			var inodoExistente *InodoStruct
			inodoExistente = new(InodoStruct)

			if CrearInodo(pathDisco, inodoExistente, sb,
				posBitmapInodo,
				arrPosicionesBD, int64(contadorPosicionesBD),
				arrPosicionesInodo, int64(contadorPosicionesInodo),
				int64(size), contenido, repetido, comando, nameProper, posOcupados) {
				return true
			} else {
				return false
			}
		} else { //no repetido
			if encontreFree != -1 { //encontre espacio libre

				var inodo *InodoStruct //----creando el inodo(archivo)
				inodo = new(InodoStruct)
				InodoInicializar(inodo, "Eduardo") //inicializado

				posBitmapInodo = int64(arrPosicionesInodo[contadorPosicionesInodo])
				contadorPosicionesInodo++

				InodoLlenar(inodo, posBitmapInodo, int64(size), -1, nameProper)

				GuardarInodo(comando, inodo, pathDisco, int(sb.SbApTablaInodo+(posBitmapInodo*sb.SbSizeStructInodo)))

				//unir el inodo con el dd
				//-----------creando el espacio en el DD, grabando info, Guardando DD actualizado
				DDAptrcrearI(dd, int(posFree), archivo, posBitmapInodo) // ingresando el dato del archivo al DD, uniendo //dd.DDArrayFile[posFree].DDFileApInodo = posBitmapInodo

				GuardarDD(comando, dd, pathDisco, int(sb.SbApDetalleDirectorio+(posBitmapDD*sb.SbSizeStructDetalleDirectorio)))
				//seguir... adentro del else superior

				posOcupados := make([]int, 0)
				if CrearInodo(pathDisco, inodo, sb,
					posBitmapInodo,
					arrPosicionesBD, int64(contadorPosicionesBD),
					arrPosicionesInodo, int64(contadorPosicionesInodo),
					int64(size), contenido, repetido, comando, nameProper, posOcupados) {
					return true
				} else {
					return false
				}

			} else { //no encontre espacio libre en los aptrs normales

				if dd.DDApDetalleDirectorio == -1 {

					var ddIndirecto *DDStruct //inicializo el indirecto uno a 0
					ddIndirecto = new(DDStruct)
					DDinicializar(ddIndirecto)

					posBitmapDDIndirecto := int64(arrPosicionesDD[contadorPosicionesDD]) //extrayendo la  posBitmap para DD
					contadorPosicionesDD++

					//uniendo el courier DD con el nuevo indirecto y guardar el courier DD
					dd.DDApDetalleDirectorio = posBitmapDDIndirecto
					GuardarDD(comando, dd, pathDisco, int(sb.SbApDetalleDirectorio+(posBitmapDD*sb.SbSizeStructDetalleDirectorio)))

					GuardarDD(comando, ddIndirecto, pathDisco, int(sb.SbApDetalleDirectorio+(posBitmapDDIndirecto*sb.SbSizeStructDetalleDirectorio)))

					//dd = ddIndirecto//ahora el courier es el indirecto
					//posBitmapDD = posBitmapDDIndirecto

					//sigo con el indirecto creado
					if mkfileCrear(pathDisco, sb, ddIndirecto, posBitmapDDIndirecto, posBitmapInodo, posBitmapBD,
						arrPosicionesBD, contadorPosicionesBD,
						arrPosicionesInodo, contadorPosicionesInodo,
						arrPosicionesDD, contadorPosicionesDD,
						archivo, size, contenido,
						encontreFree, posFree, repetido,
						comando, nameProper) {
						return true
					} else {
						return false
					}
				} else {

					//hay un indirecto existen, lo voy a buscar
					if mkfileCrear(pathDisco, sb, dd, dd.DDApDetalleDirectorio, posBitmapInodo, posBitmapBD,
						arrPosicionesBD, contadorPosicionesBD,
						arrPosicionesInodo, contadorPosicionesInodo,
						arrPosicionesDD, contadorPosicionesDD,
						archivo, size, contenido,
						encontreFree, posFree, repetido,
						comando, nameProper) {
						return true
					} else {
						return false
					}

				}
			}
		}

	} else {
		PrintError(ERROR, "existio un error al extraer el dd del archivo")
		return false
	}

}

func CrearInodo(pathDisco string, inodo *InodoStruct, sb *SuperBootStruct,
	posBitmapInodo int64, //pos courier
	arrPosicionesBD []int, contadorPosicionesBD int64, //crear nuevos BD
	arrPosicionesInodo []int, contadorPosicionesInodo int64, //crear Indirectos
	size int64, contenido string, repetido bool, comando CONSTCOMANDO, nameProper string, posOcupados []int) bool {

	posOcupados = make([]int, 0) //reincio las posiciones ocupadas

	if ExtrarInodo(pathDisco, comando, inodo, int(sb.SbApTablaInodo+(posBitmapInodo*sb.SbSizeStructInodo))) {
		PrintAviso(comando, "inodo extraido correctamente")

		for i := range inodo.IArrayBloques {
			if len(contenido) > 0 { //falta por escribir

				var bd *BloqueDeDatosStruct
				bd = new(BloqueDeDatosStruct)
				parteCortada, parteRestante := BDrecortarContenido(contenido)
				posOcupados = append(posOcupados, i)

				if inodo.IArrayBloques[i] == -1 { //espacio para crear un BD

					//inicializo//lleno
					BDinicializar(bd)
					BDllenar(bd, parteCortada)

					//extrayendo la  posBitmap para BD
					posBitmapBD := int64(arrPosicionesBD[contadorPosicionesBD])
					contadorPosicionesBD++

					//guardo el BD
					if !GuardarBD(comando, bd, pathDisco, int(sb.SbApBloques+(posBitmapBD*sb.SbSizeStructBloque))) {
						PrintAviso(comando, "no se pudo guardar el contenido del bd al inodo")
						return false
					}

					//uno el inodo con el BD
					inodo.IArrayBloques[i] = posBitmapBD //uniendo el inodo.aptr[i] con su BD

					if !GuardarInodo(comando, inodo, pathDisco, int(sb.SbApTablaInodo+(posBitmapInodo*sb.SbSizeStructInodo))) {
						PrintAviso(comando, "no se pudo ir guardando la info del bd al inodo y error al momento de escribir el inodo")
						return false
					}

				} else {

					posBitmapBD := inodo.IArrayBloques[i]

					if ExtrarBD(pathDisco, comando, bd, int(sb.SbApBloques+(posBitmapBD*sb.SbSizeStructBloque))) { //extraigo el bd existente con contenido

						BDinicializar(bd) //limpio la basura que tenia
						BDllenar(bd, parteCortada)

						if !GuardarBD(comando, bd, pathDisco, int(sb.SbApBloques+(posBitmapBD*sb.SbSizeStructBloque))) {
							PrintAviso(comando, "no se pudo guardar el contenido del bd que se sobreescribio al inodo")
							return false
						}

					} else {
						PrintError(ERROR, "error, no se pudo extraer el bd que ya tiene el archivo para sobreescribirlo")
						return false
					}

				}

				contenido = parteRestante
				inodo.ICountBloquesAsignados += int64(i + 1)

			} else { //ya se termino de escribir el archivo

				if !GuardarInodo(comando, inodo, pathDisco, int(sb.SbApTablaInodo+(posBitmapInodo*sb.SbSizeStructInodo))) {
					PrintAviso(comando, "no se pudo ir guardando la info del bd al inodo y error al momento de escribir el inodo")
					return false
				}

				PrintAviso(comando, "exitosamente Ya se termino de escribir el archivo de su contenido")
				return true
			}
		}

		//el courier ya actualizado con el contenido y bloques ocupadas hasta aqui y guardado

		if (len(posOcupados) == 4) && (len(contenido) > 0) {

			var inodoIndirecto *InodoStruct //----creando el inodo(archivo)
			inodoIndirecto = new(InodoStruct)
			InodoInicializar(inodoIndirecto, "Eduardo") //inicializado

			posBitmapInodoIndirecto := int64(arrPosicionesInodo[contadorPosicionesInodo])
			contadorPosicionesInodo++

			InodoLlenar(inodoIndirecto, posBitmapInodoIndirecto, int64(size), -1, nameProper)

			//se une el courier con el nuevo indirecto
			inodo.IApIndirecto = posBitmapInodoIndirecto

			//guarda el inodo nuevo, lleno de la info
			GuardarInodo(comando, inodoIndirecto, pathDisco, int(sb.SbApTablaInodo+(posBitmapInodoIndirecto*sb.SbSizeStructInodo)))

			//guarda el inodo courier, ya unido
			GuardarInodo(comando, inodo, pathDisco, int(sb.SbApTablaInodo+(posBitmapInodo*sb.SbSizeStructInodo)))

			if CrearInodo(pathDisco, inodoIndirecto, sb, posBitmapInodoIndirecto, arrPosicionesBD, contadorPosicionesBD, arrPosicionesInodo, contadorPosicionesInodo, size, contenido, repetido, comando, nameProper, posOcupados) {
				return true
			} else {
				return false
			}

		} else {
			PrintAviso(comando, "Se lleno el archivo existosamente (inodo de todoo su contenido)")
			return true
		}

	} else {
		PrintError(ERROR, "existio un error al extraer el inodo del archivo")
		return false
	}
}

//obtenerPosAjusteMkfile obtiene el arreglo de las posiciones que les corresponde segun la estructura en su bitmap
func obtenerPosAjusteMkfile(necesitoCount int, nameEstructura string, start int, tamanio int, comando CONSTCOMANDO, pathDisco string, nameProper string, partFit byte) (bool, []int) {

	if necesitoCount == 0 {
		PrintAviso(comando, "No necesitamos un espacio en el bitmap ya que el tamanio es igual a 0 o el necesitoCount")
		return true, nil
	}

	if necesitoCount > 0 { //dependiendo de cuantos espacios necesito ir a traer las pos con el fit

		bitmapBool, bitmapVal := ExtrarBitmap(pathDisco, comando, start, tamanio, nameEstructura)

		if bitmapBool {

			if partFit == 'B' {
				bestAjusteBool, bestAjusteVal := BitmapMejorPeorAjuste(bitmapVal, necesitoCount, comando, true, nameProper)
				if bestAjusteBool {

					for _, val := range bestAjusteVal {
						bitmapVal[val] = 1
					}
					if GuardarBitmap(comando, bitmapVal, pathDisco, start, nameEstructura) {

						return true, bestAjusteVal
					} else {
						PrintError(ERROR, "Error al momento de ocupar los espacios en el bitmap y grabarlos en la particion del disco")
						return false, nil
					}

				} else {
					PrintError(ERROR, "No fue posible calcular la posicion en el bitmap para crear con el Best Fit")
					return false, nil
				}

			} else if partFit == 'F' { //todo:posiciones relativas al bitmap
				primerAjusteBool, primerAjusteVal := BitmapPrimerAjuste(bitmapVal, necesitoCount, comando)
				if primerAjusteBool {

					for _, val := range primerAjusteVal {
						bitmapVal[val] = 1
					}
					if GuardarBitmap(comando, bitmapVal, pathDisco, start, nameEstructura) {

						return true, primerAjusteVal
					} else {
						PrintError(ERROR, "Error al momento de ocupar los espacios en el bitmap y grabarlos en la particion del disco")
						return false, nil
					}

				} else {
					PrintError(ERROR, "No fue posible calcular la posicion en el bitmap para crear con el First Fit")
					return false, nil
				}

			} else { //W
				peorAjusteBool, peorAjusteVal := BitmapMejorPeorAjuste(bitmapVal, necesitoCount, comando, false, nameProper)
				if peorAjusteBool {

					for _, val := range peorAjusteVal {
						bitmapVal[val] = 1
					}
					if GuardarBitmap(comando, bitmapVal, pathDisco, start, nameEstructura) {

						return true, peorAjusteVal
					} else {
						PrintError(ERROR, "Error al momento de ocupar los espacios en el bitmap y grabarlos en la particion del disco")
						return false, nil
					}

				} else {
					PrintError(ERROR, "No fue posible calcular la posicion en el bitmap para crear con el Worst Fit")
					return false, nil
				}

			}

		} else {
			PrintError(ERROR, "Error al extraer el bitmap correspondiente [Bitmap: "+nameEstructura+"]")
			return false, nil
		}

	} else {
		PrintAviso(comando, "Por alguna razon no se necesita ir a buscar espacio en el bitmap el numero de necesitoCount puede ser negativo")
		return false, nil
	}
}

//MkfileOpcionales evalua y retorna los opcionales
func MkfileOpcionales(isOkSize bool, valSize int, isOkCont bool, valCont string) (int, string) {
	sizeFinal := 0
	contFinal := ""

	if isOkCont && isOkSize {
		contFinal = valCont
		sizeFinal = valSize
	} else if !isOkCont && isOkSize {
		sizeFinal = valSize

		contador := 0
		for i := 0; i < sizeFinal; i++ {
			contFinal += letras[contador]

			if contador == len(letras) {
				contador = 0
				continue //TODO: validar que este parametro aqui este bien
			}

			contador++
		}
	} else if isOkCont && !isOkSize {
		contFinal = valCont
		sizeFinal = len(contFinal)
	} else { //!isOkCont && !isOkSize
		contFinal = ""
		sizeFinal = 0
	}

	//el SIZE manda
	if len(contFinal) > sizeFinal { //contenido mayor
		//cortar el contenido, hasta el size
		arrayByte := []byte(contFinal)
		arrayByteFinal := make([]byte, 0)
		for i := 0; i < sizeFinal; i++ {
			//arrayByteFinal[i] += arrayByte[i]
			arrayByteFinal = append(arrayByteFinal, arrayByte[i])
		}
		contFinal = string(arrayByteFinal)
	} else if len(contFinal) < sizeFinal { //contenido menor
		//llenar lo demas de abecedario
		//o dejarlo asi y pintar de blanco los BD pero siempre graficar los BD [*nota talvez guiarse el size y no del tamanio del contenido, para crear los BD calcular cuantos necesito]
		arrayByte := []byte(contFinal)
		arrayByteFinal := make([]byte, 0)

		arrayByteFinal = arrayByte
		//for j := 0; j < len(arrayByte); j++ {//hasta donde tengo
		//	arrayByteFinal[j] += arrayByte[j]//de lo que tengo
		//}

		contador := 0
		for i := len(arrayByteFinal); i < sizeFinal; i++ { //lo demas del abecedario
			//arrayByteFinal[i] += []byte(letras[contador])[0] //cada caracter del abecedario
			arrayByteFinal = append(arrayByteFinal, []byte(letras[contador])[0])
			if contador == len(letras)-1 {
				contador = 0
				continue
			}
			contador++
		}
		contFinal = string(arrayByteFinal)
	} else { //contenido igual al size
		//sigue normal...
	}
	return sizeFinal, contFinal
}

//ArrDir retornar el arrDir
func ArrDir(pathMkfile string) [][50]byte {
	rutadd := strings.Split(pathMkfile, "/") //"/home/user/docs/edu.txt" //4[ home user docs edu.txt]

	//arrDir[] INICALIZAR LOS NOMBRES CON [50]{00000000}
	nombreComodin := [50]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	arrDir := make([][50]byte, 0)

	//avd.AVDNombreDirectorio = nombreComodin
	for i := 0; i < len(rutadd); i++ { //a mi arreglo le ingreso la cantidad de nombres que necisitare LA INICIALIZACION
		arrDir = append(arrDir, nombreComodin) //cuantos nombres voy a necesitar
	}

	for i, val := range rutadd {
		copy(arrDir[i][:], val) //para cada nombre descompuesto lo copia en una posicion inicializada de mi arreglo de nombres inicializadas
	}

	if arrDir[0] == nombreComodin {
		copy(arrDir[0][:], "/")
	} else {
		PrintError(ERROR, "No puedes crear una carpeta afuera de la raiz")
		return nil
	}
	return arrDir
}

//DDinicializar crea un dd
func DDinicializar(dd *DDStruct) {
	fechaCreacion := [25]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	nombreComodin := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}

	for i := range dd.DDArrayFile {
		dd.DDArrayFile[i] = DDinfoStruct{}
	}

	for i := range dd.DDArrayFile { //inicializacion
		dd.DDArrayFile[i].DDFileNombre = nombreComodin
		dd.DDArrayFile[i].DDFileApInodo = -1
		dd.DDArrayFile[i].DDFileDateCreacion = fechaCreacion
		dd.DDArrayFile[i].DDFileDateModificacion = fechaCreacion
	}
	dd.DDApDetalleDirectorio = -1
}

func DDAptrmodificarI(dd *DDStruct, posicionI int, fileInodo int64) {
	dd.DDArrayFile[posicionI].DDFileApInodo = fileInodo
	copy(dd.DDArrayFile[posicionI].DDFileDateModificacion[:], time.Now().Format("01-02-2006 15:04:05"))
}

func DDAptrcrearI(dd *DDStruct, posicionI int, fileName string, posBitmapInodo int64) {
	copy(dd.DDArrayFile[posicionI].DDFileNombre[:], fileName)
	dd.DDArrayFile[posicionI].DDFileApInodo = posBitmapInodo
	copy(dd.DDArrayFile[posicionI].DDFileDateCreacion[:], time.Now().Format("01-02-2006 15:04:05"))
	copy(dd.DDArrayFile[posicionI].DDFileDateModificacion[:], time.Now().Format("01-02-2006 15:04:05"))

}

//InodoInicializar inicializa un inodo
func InodoInicializar(inodo *InodoStruct, nameProper string) {
	inodo.ICountInodo = -1
	inodo.ISizeArchivo = -1
	inodo.ICountBloquesAsignados = -1
	for i := range inodo.IArrayBloques {
		inodo.IArrayBloques[i] = -1
	}
	inodo.IApIndirecto = -1

	nombreComodin := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	inodo.IIdProper = nombreComodin
	copy(inodo.IIdProper[:], nameProper)
}

//InodoLlenar llena la infor al inodo
func InodoLlenar(inodo *InodoStruct, posBitmapInodo int64, sizeArchivo int64, apIndirecto int64, nameProper string) {
	inodo.ICountInodo = posBitmapInodo
	inodo.ISizeArchivo = sizeArchivo
	necesitoBD := math.Ceil(float64(sizeArchivo) / 25) //aprox al mayor
	inodo.ICountBloquesAsignados = int64(necesitoBD)
	inodo.IApIndirecto = apIndirecto
	copy(inodo.IIdProper[:], nameProper)
}

//BDinicializar inicialia un bd
func BDinicializar(bd *BloqueDeDatosStruct) {
	nombreComodin := [25]byte{'a', 'b', 'c', 'd', 'e', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x'}
	bd.DbData = nombreComodin
}

//BDllenar llena de contenido un bd
func BDllenar(bd *BloqueDeDatosStruct, contenido25 string) {
	copy(bd.DbData[:], contenido25)
}

//BDrecortarContenido recorta el string y te devuelve la parte cortada y la parte restante
func BDrecortarContenido(contenido string) (string, string) {
	if len(contenido) > 25 {
		arrayByteContenido := []byte(contenido)

		parteCortada := make([]byte, 0)
		for i := 0; i < 25; i++ {
			parteCortada = append(parteCortada, arrayByteContenido[i])
		}

		parteRestante := make([]byte, 0)
		for i := len(parteCortada); i < len(arrayByteContenido); i++ {
			parteRestante = append(parteRestante, arrayByteContenido[i])
		}
		return string(parteCortada), string(parteRestante)
	} else {
		return contenido, ""
	}
}

//DDcrear crea un DD inicializado y unido al AVD
func DDcrear(pathDisco string, posBitmapAvd int64, sb *SuperBootStruct, avdNew *AVDStruct, arrPosicionesDD []int, comando CONSTCOMANDO) (bool, int64) {

	if avdNew.AVDApDetalleDirectorio == -1 {

		var dd *DDStruct
		dd = new(DDStruct)

		DDinicializar(dd) //un nuevo DD

		posBitmapDD := int64(arrPosicionesDD[0]) //el unico que deberia de tener

		avdNew.AVDApDetalleDirectorio = posBitmapDD //uniendo el AVD-DD

		if GuardarAVD(comando, avdNew, pathDisco, int(sb.SbApArbolDirectorio+(posBitmapAvd*sb.SbSizeStructArbolDirectorio))) {

			if GuardarDD(comando, dd, pathDisco, int(sb.SbApDetalleDirectorio+(posBitmapDD*sb.SbSizeStructDetalleDirectorio))) {

				PrintAviso(comando, "Exito, se creo el nuevo DD, se unio con su AVD, se guardo el AVD y el DD")
				return true, posBitmapDD

			} else {
				PrintError(ERROR, "Existio un error al momento de guardar el nuevo DD del archivo que se deseaba crear")
				return false, -1
			}

		} else {
			PrintError(ERROR, "Existio un error al momento de grabar el AVD actualizado de la carpeta con la info del nuevo Detalle de Directorio")
			return false, -1
		}

	} else {
		return true, avdNew.AVDApDetalleDirectorio
	}
}

//todo: las posiciones son para grabarlas, para sabe donde grabarlas

func MkdirFinal(pathDisco string, rutaMkdir string, nameProper string, partFit byte, isP bool, comando CONSTCOMANDO, nodoDis *NodoDisco, nodoPart *NodoParticion) {

	rutadd := strings.Split(rutaMkdir, "/") //"/home/user/docs" //4[ home user docs]

	//arrDir[] INICALIZAR LOS NOMBRES CON [50]{00000000}
	nombreComodin := [50]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	arrDir := make([][50]byte, 0)

	//avd.AVDNombreDirectorio = nombreComodin
	for i := 0; i < len(rutadd); i++ { //a mi arreglo le ingreso la cantidad de nombres que necisitare LA INICIALIZACION
		arrDir = append(arrDir, nombreComodin) //cuantos nombres voy a necesitar
	}

	//copy(avd.AVDNombreDirectorio[:], nameDir)
	for i, val := range rutadd {
		copy(arrDir[i][:], val) //para cada nombre descompuesto lo copia en una posicion inicializada de mi arreglo de nombres inicializadas
	}

	if arrDir[0] == nombreComodin {
		copy(arrDir[0][:], "/")
	} else {
		PrintError(ERROR, "No puedes crear una carpeta afuera de la raiz")
		return
	}

	var avdNew *AVDStruct
	avdNew = new(AVDStruct)

	var sb *SuperBootStruct
	sb = new(SuperBootStruct)
	var sbCopia *SuperBootStruct
	sbCopia = new(SuperBootStruct)

	if ObtenerSBySBCopia(nodoDis, nodoPart, sb, sbCopia, comando) {

		contadorArr := int64(0)
		posAvdBitmapRoot := int64(0)
		countFree := 1
		firstFree := 0
		arrPosiciones := obtenerPosAjuste(pathDisco, sb, rutaMkdir, nameProper, comando, partFit)
		contadorPosiciones := 0
		if arrPosiciones != nil {

			if mkdirCrear(pathDisco, avdNew, arrDir, contadorArr, sb.SbApArbolDirectorio, posAvdBitmapRoot, sb.SbSizeStructArbolDirectorio, countFree, firstFree, comando, isP, arrPosiciones, contadorPosiciones, nameProper) {
				PrintAviso(comando, "EXCELENTE EDUARDO acabas de crear una ruta de carpetas")
				return
			} else {
				PrintError(ERROR, "Existio algun inconveniente al crear una ruta")
				return
			}

		} else {
			PrintError(ERROR, "Error al calcular las posiciones en el bitmap con el ajuste solicitado")
			return
		}

	} else {
		PrintError(ERROR, "Existio un error al extraer el SB principal de la particion")
		return
	}

}

//todo: ir a grabar el bitmap final YAAAA
//TODO: NEW ANTES?
//todo: verificar los retornos
//TODO AQUI LE AGREGUE EL INCONDICIONAL DE if int(contadorArr) < len(arrDir){, ver
//TODO SINO LO NECESITO EN OTRO LADO

func mkdirCrear(pathDisco string, avd *AVDStruct, arrDir [][50]byte, contadorArr int64, startAVDStructs int64, posAVDBitmap int64, sizeAVDStruct int64, contFree int, firstFree int, comando CONSTCOMANDO, isP bool, arrPosiciones []int, contadorPosiciones int, nameProper string) bool {

	avd = new(AVDStruct) //TODO: revisar esto porque quiero que cada nivel sea nuevo
	//FIXME: reiniciar firstFree = x //cuando empieza de nuevo pero en un nuevo avd YA
	//FIXME: :::::::::::::: contFree = 1 YA
	contFree = 1
	if ExtrarAVD(pathDisco, comando, avd, int(startAVDStructs+(posAVDBitmap*sizeAVDStruct))) { //un nuevo avd //en la posicion del Disco

		if int(contadorArr) < len(arrDir) {
			if avd.AVDNombreDirectorio == arrDir[contadorArr] {
				contadorArr++ //TODO: AVANCE encontrado el courier avanzo al siguiente

				for i, val := range avd.AVDApArraySubdirectorios {
					if val != -1 {
						//posAVDBitmap =  //TODO: AVANCE
						if mkdirCrear(pathDisco, avd, arrDir, contadorArr, startAVDStructs, val, sizeAVDStruct, contFree, firstFree, comando, isP, arrPosiciones, contadorPosiciones, nameProper) {
							PrintAviso(comando, "Si fue posible crear hasta la carpeta [Nombre: "+string(arrDir[contadorArr][:])+"]") //TODO: CONFIRMA QUE SI SEA ESTE VALOR Y NO EL AVD.NOMBRE QUIZAS PORQUE ES EL MAS RECIENTE
							return true
						}

					} else {
						if contFree == 1 { //si es el primero
							firstFree = i //TODO: AVANCE
							contFree++
						}
						//sino sigue normal, no me intersan los demas -1
					}
				}

				if avd.AVDApArbolVirtualDirectorio != -1 { //tiene valor

					//posAVDBitmap = avd.AVDApArbolVirtualDirectorio //relativa al bitmap
					//contadorArr = 0
					//FIXME: contadorArr-- YA
					contadorArr--
					if mkdirCrear(pathDisco, avd, arrDir, contadorArr, startAVDStructs, avd.AVDApArbolVirtualDirectorio, sizeAVDStruct, contFree, firstFree, comando, isP, arrPosiciones, contadorPosiciones, nameProper) {
						PrintAviso(comando, "Si fue posible crear hasta la carpeta [Nombre: "+string(arrDir[contadorArr][:])+"]") //TODO: CONFIRMA QUE SI SEA ESTE VALOR Y NO EL AVD.NOMBRE QUIZAS PORQUE ES EL MAS RECIENTE
						return true
					} else {
						PrintAviso(comando, "SE BUSCO EN TODOS LOS APTRS Y EN EL INDIRECTO PERO NO EXISTE UNA CARPETA ASI QUE SE PROCEDE A VALIDAR SI PODEMOS CREARLA, SI ES PADRE O ES POR LA QUE VENIMOS A CREAR")
						//seguir...
					}

				} else { //es == -1
					if contFree == 1 { //si es el primero, porque no habia espacio aun y todos estaban llenos, valido si tengo
						firstFree = 6 //el aptr indirecto
						contFree++
						//CREO EL STRUCT INDIRECTO PERO ANTES VALIDAR SI TENGO PERMISO DE CREAR RECURSIVO SINO NO
					}
					//seguir...
				}

				faltantes := int64(len(arrDir)) - contadorArr //es por la que vengo o son padres

				if faltantes > 1 {

					if isP {

						//TODO:CREAR CARPETA

						if firstFree == 6 {
							PrintAviso(comando, "F firstFree == 6... crear el struct indirecto primero con toda y su nueva y traida informacion y apartir de aqui crear la nueva, AQUI UNIRLA CON LA DE ADELANTE \"LA NUEVA")

							//modificando el courier con la info del indirecto
							posicionIndirectoRelativo := int64(arrPosiciones[contadorPosiciones])
							contadorPosiciones++

							avd.AVDApArbolVirtualDirectorio = posicionIndirectoRelativo                            //modificando el courier con la info relativa
							GuardarAVD(comando, avd, pathDisco, int(startAVDStructs+(posAVDBitmap*sizeAVDStruct))) //grabar el courier con la info del indirecto con la info relativa + real del disco

							//el nuevo Indirecto con la info del courier
							var avdIndirecto *AVDStruct
							avdIndirecto = new(AVDStruct)
							crearAVD(avdIndirecto, string(avd.AVDNombreDirectorio[:]), string(avd.AVDProper[:])) //el nuevo indirecto

							//modificando el nuevo indirecto con la info de la siguiente carpeta
							posicionIndirectoRelativo2 := int64(arrPosiciones[contadorPosiciones])

							avdIndirecto.AVDApArraySubdirectorios[0] = posicionIndirectoRelativo2                                        //modificando el nuevo indirecto con la info de la nueva carpeta
							GuardarAVD(comando, avdIndirecto, pathDisco, int(startAVDStructs+(posicionIndirectoRelativo*sizeAVDStruct))) //grabar el indirecto con la info de la nueva? si arriba en la linea esta ya la info de la nueva

							//avdIndirecto actualizando la info del courier
							avd = avdIndirecto
							posAVDBitmap = posicionIndirectoRelativo
							firstFree = 0
							//FIXME: ver que no tenga que aumentar al sigui. nombre de la carpeta
						}

						//con aptrs normales

						posicionRelativa := int64(arrPosiciones[contadorPosiciones]) //pos nueva
						contadorPosiciones++                                         //lista para usar el sig disponible

						avd.AVDApArraySubdirectorios[firstFree] = posicionRelativa                             //uniendo el courier con el nuevo y
						GuardarAVD(comando, avd, pathDisco, int(startAVDStructs+(posAVDBitmap*sizeAVDStruct))) //pos courier//grabando de nuevo el courier

						//crear la nueva carpeta por la que vengo unirla al courier
						var avdNuevaCarpeta *AVDStruct
						avdNuevaCarpeta = new(AVDStruct)
						crearAVD(avdNuevaCarpeta, string(arrDir[contadorArr][:]), nameProper)
						GuardarAVD(comando, avdNuevaCarpeta, pathDisco, int(startAVDStructs+(posicionRelativa*sizeAVDStruct))) //grabar el nuevo

						//todo????aqui no le asigno al home el aptr de la otra carpeta? mas adelante?
						//actualizando el siguiente a extraer
						//posAVDBitmap =  //la direccion del que acabas de crear, para ir a extraer este

						//y seguir de ser necesario
						contadorArr++

						if contadorArr == int64(len(arrDir)) {
							PrintAviso(comando, "ya fue creado el PATH completo")
							return true
						} else {
							contadorArr--
							if mkdirCrear(pathDisco, avd, arrDir, contadorArr, startAVDStructs, posicionRelativa, sizeAVDStruct, contFree, firstFree, comando, isP, arrPosiciones, contadorPosiciones, nameProper) {
								PrintAviso(comando, "Si fue posible crear hasta la carpeta [Nombre: "+string(arrDir[contadorArr][:])+"]") //TODO: CONFIRMA QUE SI SEA ESTE VALOR Y NO EL AVD.NOMBRE QUIZAS PORQUE ES EL MAS RECIENTE
								return true
							} else {
								PrintAviso(comando, "No se puede crear una carpeta ************ verificar este fin ")
								return false
								//seguir...
							}
						}

					} else {
						PrintAviso(comando, "La carpeta se identifico que es padre y No se puede crear porque [Dir: "+string(arrDir[contadorArr][:])+"arrDir[contadorArr] no existe y no se puede crear a falta del parametro -P ....")
						return false
					}

				} else {
					if faltantes == 1 { //por la que vine

						//TODO:CREAR CARPETA

						//con el indirecto

						if firstFree == 6 {
							PrintAviso(comando, "F firstFree == 6... crear el struct indirecto primero con toda y su nueva y traida informacion y apartir de aqui crear la nueva, AQUI UNIRLA CON LA DE ADELANTE \"LA NUEVA")

							//modificando el courier con la info del indirecto
							posicionIndirectoRelativo := int64(arrPosiciones[contadorPosiciones])
							contadorPosiciones++

							avd.AVDApArbolVirtualDirectorio = posicionIndirectoRelativo                            //modificando el courier con la info relativa
							GuardarAVD(comando, avd, pathDisco, int(startAVDStructs+(posAVDBitmap*sizeAVDStruct))) //grabar el courier con la info del indirecto con la info relativa + real del disco

							//el nuevo Indirecto con la info del courier
							var avdIndirecto *AVDStruct
							avdIndirecto = new(AVDStruct)
							crearAVD(avdIndirecto, string(avd.AVDNombreDirectorio[:]), string(avd.AVDProper[:])) //el nuevo indirecto

							//modificando el nuevo indirecto con la info de la siguiente carpeta
							posicionIndirectoRelativo2 := int64(arrPosiciones[contadorPosiciones])

							avdIndirecto.AVDApArraySubdirectorios[0] = posicionIndirectoRelativo2                                        //modificando el nuevo indirecto con la info de la nueva carpeta
							GuardarAVD(comando, avdIndirecto, pathDisco, int(startAVDStructs+(posicionIndirectoRelativo*sizeAVDStruct))) //grabar el indirecto con la info de la nueva? si arriba en la linea esta ya la info de la nueva

							//avdIndirecto actualizando la info del courier
							avd = avdIndirecto
							posAVDBitmap = posicionIndirectoRelativo
							firstFree = 0
							//FIXME: ver que no tenga que aumentar al sigui. nombre de la carpeta
						}

						//con aptrs normales

						posicionRelativa := int64(arrPosiciones[contadorPosiciones]) //pos nueva
						contadorPosiciones++                                         //lista para usar el sig disponible
						//todo: ??seria bueno que pasa si ya no tengo espacios? pasaria eso?
						avd.AVDApArraySubdirectorios[firstFree] = posicionRelativa                             //uniendo el courier con el nuevo y
						GuardarAVD(comando, avd, pathDisco, int(startAVDStructs+(posAVDBitmap*sizeAVDStruct))) //pos courier//grabando de nuevo el courier

						//crear la nueva carpeta por la que vengo unirla al courier
						var avdNuevaCarpeta *AVDStruct
						avdNuevaCarpeta = new(AVDStruct)
						crearAVD(avdNuevaCarpeta, string(arrDir[contadorArr][:]), nameProper)
						GuardarAVD(comando, avdNuevaCarpeta, pathDisco, int(startAVDStructs+(posicionRelativa*sizeAVDStruct))) //grabar el nuevo

						//actualizando el siguiente a extraer
						//posAVDBitmap = posicionRelativa //la direccion del que acabas de crear, para ir a extraer este

						//y seguir de ser necesario
						contadorArr++

						if contadorArr == int64(len(arrDir)) { //AQUI le pongo fin
							PrintAviso(comando, "ya fue creado el PATH completo")
							return true
						} else { //FIXME: esto es innecesario, al menos aqui por la resta que esta arriba sabemos que sera siempre el ultimo
							contadorArr--
							if mkdirCrear(pathDisco, avd, arrDir, contadorArr, startAVDStructs, posicionRelativa, sizeAVDStruct, contFree, firstFree, comando, isP, arrPosiciones, contadorPosiciones, nameProper) {
								PrintAviso(comando, "Si fue posible crear hasta la carpeta [Nombre: "+string(arrDir[contadorArr][:])+"]") //TODO: CONFIRMA QUE SI SEA ESTE VALOR Y NO EL AVD.NOMBRE QUIZAS PORQUE ES EL MAS RECIENTE
								return true
							} else {
								PrintAviso(comando, "No se puede crear una carpeta ************ verificar este fin ")
								return false
								//seguir...
							}
						}

					} else {
						PrintAviso(comando, "La ruta ya existe amigo")
						return true
					}
				}

			} else {
				PrintAviso(comando, "No existe la carpeta de nombre [Nombre: "+string(arrDir[contadorArr][:])+"")
				return false
			}
		} else {
			return false
		}

	} else {
		PrintError(ERROR, "Existio un error al extraer el avd del disco")
		return false
	}
}

//TODO: crear, guardar, modificar disco, log, sb, sb copia?

//obtenerAjuste obtiene las posiciones relativas al bitmap y no a la posicion del disco dependiendo el ajuste y GRABA EL BITMAP EN EL DISCO
func obtenerPosAjuste(pathDisco string, sb *SuperBootStruct, rutaMkdir string, nameProper string, comando CONSTCOMANDO, partFit byte) []int {

	necesitoCount := countNotExistDir(pathDisco, sb, rutaMkdir, nameProper, comando)

	if necesitoCount == -1 {
		PrintError(ERROR, "Error al calcular los espacios para la carpetas")
		return nil
	}

	if necesitoCount != 0 { //dependiendo de cuantos espacios necesito ir a traer las pos con el fit

		bitmapBool, bitmapVal := ExtrarBitmap(pathDisco, comando, int(sb.SbApBitMapArbolDirectorio), int(sb.SbArbolVirtualCount), "AVD")

		if bitmapBool {

			if partFit == 'B' {
				bestAjusteBool, bestAjusteVal := BitmapMejorPeorAjuste(bitmapVal, necesitoCount, comando, true, nameProper)
				if bestAjusteBool {

					for _, val := range bestAjusteVal {
						bitmapVal[val] = 1
					}
					if GuardarBitmap(comando, bitmapVal, pathDisco, int(sb.SbApBitMapArbolDirectorio), "AVD") {

						return bestAjusteVal
					} else {
						PrintError(ERROR, "Error al momento de ocupar los espacios en el bitmap y grabarlos en la particion del disco")
						return nil
					}

				} else {
					PrintError(ERROR, "No fue posible calcular la posicion en el bitmap para crear una nueva carpeta con el Best Fit")
					return nil
				}

			} else if partFit == 'F' { //todo:posiciones relativas al bitmap
				primerAjusteBool, primerAjusteVal := BitmapPrimerAjuste(bitmapVal, necesitoCount, comando)
				if primerAjusteBool {

					for _, val := range primerAjusteVal {
						bitmapVal[val] = 1
					}
					if GuardarBitmap(comando, bitmapVal, pathDisco, int(sb.SbApBitMapArbolDirectorio), "AVD") {

						return primerAjusteVal
					} else {
						PrintError(ERROR, "Error al momento de ocupar los espacios en el bitmap y grabarlos en la particion del disco")
						return nil
					}

				} else {
					PrintError(ERROR, "No fue posible calcular la posicion en el bitmap para crear una nueva carpeta con el First Fit")
					return nil
				}

			} else { //W
				peorAjusteBool, peorAjusteVal := BitmapMejorPeorAjuste(bitmapVal, necesitoCount, comando, false, nameProper)
				if peorAjusteBool {

					for _, val := range peorAjusteVal {
						bitmapVal[val] = 1
					}
					if GuardarBitmap(comando, bitmapVal, pathDisco, int(sb.SbApBitMapArbolDirectorio), "AVD") {

						return peorAjusteVal
					} else {
						PrintError(ERROR, "Error al momento de ocupar los espacios en el bitmap y grabarlos en la particion del disco")
						return nil
					}

				} else {
					PrintError(ERROR, "No fue posible calcular la posicion en el bitmap para crear una nueva carpeta con el Worst Fit")
					return nil
				}

			}

		} else {
			PrintError(ERROR, "Error al extraer el bitmap correspondiente")
			return nil
		}

	} else {
		PrintAviso(comando, "El path ya existe, revisar eso porfavor.")
		PrintAviso(comando, "Por alguna razon se recorrio el avd y no hace falta crear ninguna carpeta ya todas existen, revisar eso porfavor.")
		return nil
	}
}

//countNotExistDir con el metodo recorrer hago el calculo de cuantos me faltan
func countNotExistDir(pathDisco string, sb *SuperBootStruct, rutaMkdir string, nameProper string, comando CONSTCOMANDO) int {

	rutadd := strings.Split(rutaMkdir, "/") //"/home/user/docs" //4[ home user docs]

	//arrDir[] INICALIZAR LOS NOMBRES CON [50]{00000000}
	nombreComodin := [50]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	arrDir := make([][50]byte, 0)

	//avd.AVDNombreDirectorio = nombreComodin
	for i := 0; i < len(rutadd); i++ { //a mi arreglo le ingreso la cantidad de nombres que necisitare LA INICIALIZACION
		arrDir = append(arrDir, nombreComodin) //cuantos nombres voy a necesitar
	}

	//copy(avd.AVDNombreDirectorio[:], nameDir)
	for i, val := range rutadd {
		copy(arrDir[i][:], val) //para cada nombre descompuesto lo copia en una posicion inicializada de mi arreglo de nombres inicializadas
	}

	if arrDir[0] == nombreComodin {
		copy(arrDir[0][:], "/")
	} else {
		PrintError(ERROR, "No puedes crear una carpeta afuera de la raiz")
		return -1
	}

	//return total de la resta
	var avd *AVDStruct
	avd = new(AVDStruct) //TODO: VER CUAL ME QUEDO GRABADO? EL ULTIMO O EL PRIMERO? :D jaja
	contadorArr := int64(0)
	posAVDBitmap := int64(0)
	_, valExisten, valIndirectos, _ := MkdirRecorrer(pathDisco, avd, arrDir, contadorArr, sb.SbApArbolDirectorio, posAVDBitmap, sb.SbSizeStructArbolDirectorio, 0, 1, comando)

	sizeArr := int64(len(arrDir)) //menos la raiz '/'
	faltantes := sizeArr - valExisten
	fmt.Println("Faltan estos aptrs  ,+  indirectos:::::::::::::::::::")
	fmt.Println(faltantes)
	fmt.Println(valIndirectos)
	return int(faltantes + valIndirectos) //por lo menos tiene que se mayor > 0

}

//TODO AQUI LE AGREGUE contadorArr-- cuando se metia en el aptr indirecto

//MkdirRecorrer es como un recorrer pero me va contando cuantos Dir ya existen
func MkdirRecorrer(path string, avd *AVDStruct, arrDir [][50]byte, contadorArr int64, startAVDStructs int64, posAVDBitmap int64, sizeAVDStruct int64, countIndirectos int64, countFree int, comando CONSTCOMANDO) (bool, int64, int64, int64) {
	avd = new(AVDStruct)
	countFree = 1

	if ExtrarAVD(path, comando, avd, int(startAVDStructs+(posAVDBitmap*sizeAVDStruct))) {

		if int(contadorArr) < len(arrDir) {

			if avd.AVDNombreDirectorio == arrDir[contadorArr] {
				contadorArr++

				if int(contadorArr) != len(arrDir) { //sino

					for _, val := range avd.AVDApArraySubdirectorios {
						if val != -1 {
							//posAVDBitmap = val //actualizo al siguiente avd
							if valBool, valInt, valCountIndirectos, valPos := MkdirRecorrer(path, avd, arrDir, contadorArr, startAVDStructs, val, sizeAVDStruct, countIndirectos, countFree, comando); valBool { //TODO: NO ME IMPORTA QUE SE SOBRE ESCRIBA ESTE *AVD
								PrintAviso(comando, "Ya fue encontrado el PATH completo") //TODO: REVISAR ESTE MSN
								if valInt > contadorArr {
									return true, valInt, valCountIndirectos, valPos
								}
								return true, contadorArr, valCountIndirectos, valPos
							}
						} else {
							if countFree == 1 {
								countFree++ //si tengo disponible
							}
						}
					}

					//TODO: AQUI ES DONDE VALIDAR QUE NO SIGUE SI YA ME RETORNO EL TRUE
					if avd.AVDApArbolVirtualDirectorio != -1 { //si tiene #
						//posAVDBitmap = avd.AVDApArbolVirtualDirectorio
						//contadorArr = 0
						contadorArr--
						if valBool, valInt, valCountIndirectos, valPos := MkdirRecorrer(path, avd, arrDir, contadorArr, startAVDStructs, avd.AVDApArbolVirtualDirectorio, sizeAVDStruct, countIndirectos, countFree, comando); valBool {
							PrintAviso(comando, "Ya fue encontrado el PATH completo en el aptr indirecto") //TODO: REVISAR ESTE MSN
							PrintAviso(comando, "ya fue encontrado el PATH completo")
							if valInt > contadorArr {
								return true, valInt, valCountIndirectos, valPos
							}
							return true, contadorArr, valCountIndirectos, valPos
						}
						return false, contadorArr, countIndirectos, -1 //false

					} else { //== -1
						PrintAviso(comando, "ya no existe mas aptrs para buscar ni el indirecto")
						PrintAviso(comando, "no existe ese nombre de carpeta ")
						if countFree == 1 {
							countFree++
							countIndirectos++ //necesito crear otro nodo para este avd que ya tiene todos sus aptrs ocupados
						}
						return true, contadorArr, countIndirectos, -1 //false
					}

				} else { //==len()
					PrintAviso(comando, "Ya fue encontrado el PATH completo")
					return true, contadorArr, countIndirectos, posAVDBitmap
				}

			} else {
				return false, contadorArr, countIndirectos, -1 //false
			}

		} else {
			return false, contadorArr, countIndirectos, -1 //false
		}

	} else {
		PrintError(ERROR, "Error al extraer el [AVD: Raiz]")
		return false, contadorArr, countIndirectos, -1
	}
}

//todo: tomar en cuenta que las posiciones retornadas son relativas al bitmap es decir puede ser pos == 1,2,3
//todo: ya entonces ese valor lo busco apartir de la posicion del bitmap en el disco o
//todo: lo busco desde el inicio del avd (estructuras) multiplico y grabo en ese posicion
//todo: las posiciones son relativas PERO
//todo: tomar en cuenta que a estos le tengo que + 1

//BitmapPrimerAjuste retornar un arreglo con las posiciones para el Primer ajuste
func BitmapPrimerAjuste(bitmap []byte, necesitoCount int, comando CONSTCOMANDO) (bool, []int) {

	libres := 0
	arrFrees := make([]int, 0)

	for i := 0; i < len(bitmap); i++ { //todo: aqui analizis para multiplicar  o mas desde el inicio + este val, me explico
		if bitmap[i] != 1 { //el contendio
			libres++
			arrFrees = append(arrFrees, i) //la posicion relativa 0,1,2...
			if libres == necesitoCount {
				PrintAviso(comando, "Ya SE ENCONTRO espacio en el bitmap para crear con el Primer Ajuste")
				return true, arrFrees
			}
		} else {
			libres = 0
			arrFrees = nil
			arrFrees = make([]int, 0)
			//eliminarElementos(arrFrees)
		}
	}

	PrintAviso(comando, "Ya no hay espacio en el bitmap para crear con el Primer Ajuste")
	return false, nil
}

//BitmapMejorPeorAjuste obtiene el primer o peor ajuste, retorna el arr con las posiciones del bitmap, listas para usarlas y grabarlas
func BitmapMejorPeorAjuste(bitmap []byte, necesitoCount int, comando CONSTCOMANDO, isMejorAjuste bool, nameCreacion string) (bool, []int) {

	libres := 0
	arrFrees := make([]int, 0)
	arrLista := make([][]int, 0)

	for i := 0; i < len(bitmap); i++ {
		if bitmap[i] != 1 { //==0
			libres++
			arrFrees = append(arrFrees, i) //agrega la posicion en el disco
		} else { //==1
			if libres > 0 {
				arrLista = append(arrLista, arrFrees)
				libres = 0
				arrFrees = nil
				arrFrees = make([]int, 0)
			}
		}
	}

	if libres > 0 {
		arrLista = append(arrLista, arrFrees)
		libres = 0
		arrFrees = nil
		arrFrees = make([]int, 0)
	}

	listaNueva := make([][]int, 0) //las que cumplen con >=
	for _, val := range arrLista {
		if len(val) >= necesitoCount {
			listaNueva = append(listaNueva, val)
		}
	}

	if len(listaNueva) == 0 { //no hay listas candidatas
		if isMejorAjuste {
			PrintAviso(comando, "Ya no hay espacio en el bitmap para crear "+nameCreacion+" con el Mejor Ajuste")
			return false, nil
		} else {
			PrintAviso(comando, "Ya no hay espacio en el bitmap para crear "+nameCreacion+" con el Peor Ajuste")
			return false, nil
		}
	} else { //si hay listas candidatas
		var max, min int = len(listaNueva[len(listaNueva)-1]), len(listaNueva[len(listaNueva)-1]) //tamanio del ultimo arr

		var arrMin, arrMax []int = listaNueva[len(listaNueva)-1], listaNueva[len(listaNueva)-1] //ultimo arr

		for i := len(listaNueva) - 1; i >= 0; i-- { //recorrido de retroceso por si necesito el primer ingresado
			if max < len(listaNueva[i]) { //la nueva es mayor
				max = len(listaNueva[i])
				arrMax = listaNueva[i] //un arreglo
			}
			if min > len(listaNueva[i]) { //la nueva es menor
				min = len(listaNueva[i])
				arrMin = listaNueva[i] //un arreglo
			}
		}

		posicionesLimpias := make([]int, 0)
		if isMejorAjuste { //si es el primer ajuste//el len() mas pequeno
			//for _, val := range arrMin {
			//	if val != 0{
			//		posicionesLimpias = append(posicionesLimpias, val)
			//	}
			//}//TODO: VER QUE SI ESTA BIEN EL FOR CON EL necesitoCount
			//TODO: QUE LAS POS ESTEN AL PRINCIPIO VERIFICAR ESO Y NO EN MEDIO O AL FINAL
			for i := 0; i < necesitoCount; i++ {
				posicionesLimpias = append(posicionesLimpias, arrMin[i])
			}

			PrintAviso(comando, "Ya SE ENCONTRO espacio en el bitmap para crear "+nameCreacion+" con el Mejor Ajuste [Cantidad de Posiciones: "+strconv.Itoa(len(posicionesLimpias))+"]")
			return true, posicionesLimpias
		} else { //si es el peor ajuste//el len() mas grande
			for i := 0; i < necesitoCount; i++ {
				posicionesLimpias = append(posicionesLimpias, arrMax[i])
			}
			PrintAviso(comando, "Ya SE ENCONTRO espacio en el bitmap para crear "+nameCreacion+" con el Peor Ajuste [Cantidad de Posiciones: "+strconv.Itoa(len(posicionesLimpias))+"]")
			return true, posicionesLimpias
		}
	}
}

//TODO: DE SEGURO IRE AGREGANDO MAS VALIDACIONES

//ValidacionParticion validaciones de mkdir antes de ejecutar
func ValidacionParticion(nodoDis *NodoDisco, nodoPart *NodoParticion, comando CONSTCOMANDO) bool {

	if nodoPart.isPartFormatLWH { //si la particion ya esta formateada

		if nodoPart.islogeado { //si estas logeado

			return true

		} else { //no estas logeado
			PrintError(ERROR, "Error no hay ninguna sesion iniciada, necesitas estar logeado")
			return false
		}

	} else { //no esta formateada
		PrintError(ERROR, "Esta particion no cuenta con el sistema de archivos LWH, no ha sido formateada")
		return false
	}

}

//ObtenerSBySBCopia obtiene el SB principal y el SB copia
func ObtenerSBySBCopia(nodoDis *NodoDisco, nodoPart *NodoParticion, sb *SuperBootStruct, sbCopia *SuperBootStruct, comando CONSTCOMANDO) bool {
	pathDisco := nodoDis.path
	namePart := nodoPart.nombre

	isOk, startPartition, _, _ := StartSizeParticion(nodoPart)

	if isOk { //informacion correcta de la particion montada
		if ExtrarSB(pathDisco, comando, sb, startPartition) { //extraer el SB

			PrintAviso(comando, "Se extajo el SB de la particion del disco [Disco :"+pathDisco+", Particion: "+namePart+"]")

			if ExtrarSB(pathDisco, comando, sbCopia, int(sb.SBApSBCopy)) { //extraer el SB copia
				PrintAviso(comando, "Se extajo el SB Copia de la particion del disco [Disco :"+pathDisco+", Particion: "+namePart+"]")
			} else {
				PrintError(ERROR, "Error al extraer el SB Copia de la particion del disco, igual se continua pero no se guardar la informacion para el recovery [Disco :"+pathDisco+", Particion: "+namePart+"]")
			}

			return true //no importa si solo el SB principal se extrajo

		} else {
			PrintError(ERROR, "Error al extraer el SB de la particion del disco [Disco :"+pathDisco+", Particion: "+namePart+", Posicion del SB: "+strconv.Itoa(startPartition)+"]")
			return false
		}
	} else {
		PrintError(ERROR, "Error al extraer informacion de la particion montada")
		return false
	}
}

//TODO: SIEMPRE VERIFICAR QUE SI TOME BIEN EL INICIO DEL EBR O PARTICION

//recorridoMkdir recorre las carpetas hasta encontrar una con el mismo nombre
//func recorridoMkdir(nodoDisco *NodoDisco, nodoParticion *NodoParticion, sb *SuperBootStruct, nameDir string, corriente int, comando CONSTCOMANDO){
//
//
//	pathDisco := nodoDisco.path
//	arrPathMkdir := strings.Split(nameDir, "/")
//	byteStartAVD := sb.SbApArbolDirectorio
//
//	nameDir := [50]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
//	copy(nameDir[:], arrPathMkdir[corriente])
//
//	var avdRaiz *AVDStruct
//	avdRaiz = new(AVDStruct)
//
//	if ExtrarAVD(pathDisco,comando, avdRaiz, int(byteStartAVD)){//extraigo la raiz
//
//		for i, val := range avdRaiz.AVDApArraySubdirectorios{
//
//			var avdCarrier *AVDStruct
//			avdCarrier = new(AVDStruct)
//
//			if val != -1{//esta usado tiene una direccion
//
//				if ExtrarAVD(pathDisco, comando, avdCarrier, int(byteStartAVD) + int(val)*binary.Size(AVDStruct{})){//si pudo leer el avd courier
//
//					if avdCarrier.AVDNombreDirectorio == pathDir{//si tiene el mismo nombre
//						//ir a buscar entre sus apuntadores del corriente la siguiente porcion si existiera
//					}else{//no tiene el mismo nombre
//						//continuar recursivamente
//						corriente++
//						//recorridoMkdir()
//					}
//
//				}else{//no se pudo extraer por algun error de lectura
//
//				}
//
//			}else{//no tiene una direccion
//				if avdRaiz.AVDApArbolVirtualDirectorio != -1{//ver si tiene un apuntador indirecto
//
//				}else{//no tiene un aptr indirecto
//
//				}
//			}
//
//		}
//
//	}else{
//		PrintError(ERROR, "Error al extraer el primer avd es decir la raiz")
//		return
//	}
//
//
//}

//StartSizeParticion retorna el inicio de la particion sin importar el tipo, [correcto,start,size,isEBR]
func StartSizeParticion(nodoPart *NodoParticion) (bool, int, int, bool) {

	var partition *PartitionStruct
	partition = nodoPart.partition

	var ebr *EBRStruct
	ebr = nodoPart.ebr

	tamanioParticion := 0
	partStart := 0
	isEBR := false

	if partition != nil { //si es primaria
		tamanioParticion = int(partition.PartSize)
		partStart = int(partition.PartStart)
		isEBR = false
	} else if ebr != nil { //si es logica
		tamanioParticion = int(ebr.PartSize) - binary.Size(ebr) //esto porque incluyo el tamanio del EBR en el tamanio de la particion
		partStart = int(ebr.PartStart) + binary.Size(ebr)       //esto porque incluyo el tamanio del EBR en el tamanio de la particion
		isEBR = true
	} else { //en mkfs no se puede formatear otro tipo de particion
		PrintError(ERROR, "Error, no se puede calcular el inicio o el tamanio para particion que no sea Primario o Logica")
		return false, -1, -1, false
	}
	return true, partStart, tamanioParticion, isEBR
}

//partFitParticion retorna el fit de una particion independiente si es Primaria o Logica
func partFitParticion(nodoPart *NodoParticion) (bool, byte) {

	var partition *PartitionStruct
	partition = nodoPart.partition

	var ebr *EBRStruct
	ebr = nodoPart.ebr

	var fit byte

	if partition != nil { //si es primaria
		fit = partition.PartFit
	} else if ebr != nil { //si es logica
		fit = ebr.PartFit
	} else { //en mkfs no se puede formatear otro tipo de particion
		PrintError(ERROR, "Error, no se puede calcular el fit para particion que no sea Primario o Logica")
		return false, fit
	}
	return true, fit
}

//===================================================================================================================LWH
//ICountInodo            int64    //numero de i-nodo
//ISizeArchivo           int64    //tamanio del archivo
//ICountBloquesAsignados int64    //numero de bloques asignados
//IArrayBloques          [4]int64 //arreglo de 4 aputandores a bloques de datos para guardar el archivo
//IApIndirecto           int64    //un apuntador indirecto por si el archivo ocupa mas de 4 bloques de datos, para el manejo de archivos de tamanio "grande"
//IIdProper				 [16]byte //identificador del propietario del archivo

//Mkfile crea un archivo con datos de uno o mas Bitmap y uno o mas BD
func Mkfile(nodoDisco *NodoDisco, nodoParticion *NodoParticion, numberInodo int64, sizeFile int64) {

	//CONSULTAR EN BITMAPS QUE NUMERO TOCA
	//CONSULTA EN EL BITMAP DE DD(SOLO UN NUMERO PORQUE SOLO CREO UN ARCHIVO)(TENGO QUE
	//PREGUNTAR SI YA EXISTE UNO, SI ES ASI, VER SI TODAVIA LE CABE OTRO ARCHIVO O CREO OTRO APUNTADOR
	//INDIRECTO)(DEPENDIENDO SI ES EN LA MISMA RUTA O CARPETA "RUTA", SI ES OTRA VALIDAR SI YA TIENE Y LO MISMO SINO CREO
	//UN APUNTADOR INDERECTO Y SIGO, SI NO TIENE LA RUTA UN DD CREO UNO), EN EL BITMAP DE INODO(DEPENDIENDO DEL SIZE DEL ARCHIVO)
	//CREO EL STRUCT DEL DD CON LA INFO DEL ARCHIVO
	//LO GRABO
	//--------
	//RECIBO EL NUMERO DE INODO PARA CONSTRUIR OTRO INODO
	//CALCULO NUMEROS (CUANTOS BD NECESITO)
	//CONSULTAR EL BITMAP DE BD (DEPENDIENDO DEL SIZE DEL ARCHIVO TAMBIEN)
	//CREO OTRO INODO//DEPENDIENDO SUS//BD TAMBIEN
	//GUARDAR ESTRUCTURAS
	//IR A ESCRIBIR AL BITMAP O ARRRIBA
	//MODIFICAR EL SB Y EL SB COPIA

	//numberBDAsigned := (sizeFile)numeroCaracteres o contendioFile / 25
}

func MkdirEjecutar(rutaMkdir string, nameProper string) {
	//descomponer la ruta
	//rutadd := strings.Split(rutaMkdir, "/") //"/home/user/docs" //4[ home user docs]
	//Mkdir(rutadd, "", 0, int64(len(rutadd)), nameProper)//paso el arreglo de los directorios

}

//TODO: TOMAR EN CUENTA QUE SOLO LA RUTA TENGO QUE PASAR SIN EL NOMBRE.TXT

//Mkdir crea carpetas dependiendo de la ruta
//func Mkdir(dir []string, completando string, inicio int64, final int64, nameProper string){
//
//	if inicio == 0{
//		//raiz
//	}else {
//		//dir1//dir2//dir3
//	}
//
//	//CONSULTAR EN EL BITMAP QUE NUMERO TOCA DEPENDIENDO TODA LA RUTA COMPLETA ES DECIR DEPENDIENDO DEL
//	//# QUE ME DIO ARRIBA DESCOMPONIENDOLO
//	//CON ESE NUMERO BUSCAR CON EL FIT EL MEJOR/PEOR/PRIMER ESPACIO DONDE ME CABEN Y JALAR ESOS NUMEROS DISPONIBLES DEL BITMAP
//	//ME RETORNAN 3 NUMEROS O EL MISMO #NUMERO CALCULADO ARRIBA DE
//	//CON ESOS NUMEROS CONSTRUIR LOS STRUCTS
//
//	//INICIA CON EL PRIMER NUMERO
//	//SIGUE SUMANDO
//	//HASTA LLEGAR AL FINAL IF
//
//	//contador de aptr ir sumando y asi hasta llegar al tipo y crear el indirecto de ser necesario?pero esto cuando veo que en un struct ya no cabe
//	//entonces ir a validar siempre si en el struct quepo..el struct de la carpeta
//	//crear un nuevo struct
//	//ir a guardar ir a extraer o modifica y luego ir a guardar? en que momento se guardan
//	//ir verificando si todavia tengo espacios apt sino crear uno indirecto y continuar con el anterior o el indirecto//mismo proc ir a traer la info del bitmap y a seguir
//
//
//	var avd *AVDStruct
//	avd = new(AVDStruct)
//	//if si tiene espacio en el struct
//	//sino creo uno nuevo
//	crearAVDstruct(avd, nameDir, nameProper)
//
//
//	if inicio != final{//sino llegas sigues creando una estructura con la info de la anterior
//		inicio++
//		completando +=  dir[inicio] + "/"
//		Mkdir(dir, completando, inicio, final, nameProper)
//	}
//	return //termina
//
//}

//AVDFechaCreacion            [25]byte
//AVDNombreDirectorio         [25]byte
//AVDApArraySubdirectorios    [6]int64 //arreglo de apuntadores directos a sub-directorios
//AVDApDetalleDirectorio      int64    //un apuntador a un detalle de directorio. Este solo se utilizara en el primer directorio
//AVDApArbolVirtualDirectorio int64    //un apuntador a otro mismo tipo de estructura por si se usan los 6 apuntadores del arreglo de sub-directorios para que puedan seguir creciendo los subdirectorios
//AVDProper                   [16]byte //id del propietario de la carpeta, el que se a generado al momento de

//crearAVD crea un struct nuevo de avd con -1 sus enteros solo el nameDir y el nameProper modificado
func crearAVD(avd *AVDStruct, nameDir string, namePropietario string) {

	fechaCreacion := [25]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	avd.AVDFechaCreacion = fechaCreacion //fecha de creacion del sistema
	copy(avd.AVDFechaCreacion[:], time.Now().Format("01-02-2006 15:04:05"))

	nombreComodin := [50]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	avd.AVDNombreDirectorio = nombreComodin
	copy(avd.AVDNombreDirectorio[:], nameDir)

	for i := range avd.AVDApArraySubdirectorios {
		avd.AVDApArraySubdirectorios[i] = -1
	}

	avd.AVDApDetalleDirectorio = -1

	avd.AVDApArbolVirtualDirectorio = -1

	avd.AVDProper = nombreComodin
	copy(avd.AVDProper[:], namePropietario)
}

//ParametroPathMkdir scanner para -path
func ParametroPathMkdir(comando CONSTCOMANDO, mapa map[string]string) int {
	if val, ok := mapa["PATH"]; ok {
		if val != "" {
			return 0
		}
		PrintError(comando, "El valor del parametro path es vacio ["+val+"]")
		return -1
	}
	PrintAviso(comando, "El parametro obligatorio path no esta en la sentencia...")
	return -1
}

//ParametroRutaRep scanner para -ruta
func ParametroRutaRep(comando CONSTCOMANDO, mapa map[string]string) bool {
	if val, ok := mapa["RUTA"]; ok {
		if val != "" {
			return true
		}
		PrintError(comando, "El valor del parametro ruta es vacio ["+val+"]")
		return false
	}
	PrintAviso(comando, "El parametro opcional ruta no esta en la sentencia...")
	return false
}

//ParametroisP scanner para -P
func ParametroisP(comando CONSTCOMANDO, mapa map[string]string) bool {
	if val, ok := mapa["P"]; ok {
		if val == "P" {
			PrintError(comando, "El parametro opcional -p es correcto ["+val+"]")
			return true
		}
		PrintError(comando, "El parametro -p tiene un valor extrano ["+val+"]")
		return false
	}
	PrintAviso(comando, "El parametro opcional -p no esta en la sentencia...")
	return false
}

//ParametroCont scanner para cont
func ParametroCont(comando CONSTCOMANDO, mapa map[string]string) bool {
	if val, ok := mapa["CONT"]; ok {
		if val == "" {
			PrintError(comando, "El valor para el parametro cont es vacio ["+val+"]")
			return true
		}
		PrintError(comando, "El parametro opcional cont se encontro dentro de la sentencia ["+val+"]")
		return true
	}
	PrintAviso(comando, "El parametro opcional cont no esta en la sentencia...")
	return false
}

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++MKFS
//hago formateos en las particiones *Partition, *EBR
//FORMATEAR PARTICION LWF//CALCULAR NUMERO DE ESTRUCTURAS//CALCULAR INICION Y TAMANIOS//CREAR SB//GUARDAR SB
//IR A ESCRIBIR LOS BITMAPS//GUARDAR SB COPIA //TODO:MODIFICAR EL SB COPIA TAMBIEN
//CREAR EL DIRECTORIO LA RAIZ
//CREAR EL ARCHIVO [/users.txt]//LLENARLO DE INFORMACION//USUARIO Y CONTRASENA
//TODO:MODIFICAR EL SB COPIA TAMBIEN

//ComandoMKFS ejecuta el comando MKFs
func ComandoMKFS(nodoDis *NodoDisco, nodoPart *NodoParticion, tipoFormateo string, nameProper string, comando CONSTCOMANDO, cantidadLWH *int64) {
	//DEPENDIENDO EL TIPO DE FORMATEO IR A FORMATEAR ANTES CON UN IF
	//YA DESPUES DARLE EL FORMAO LWH

	var sb *SuperBootStruct
	sb = new(SuperBootStruct)

	if formatearLWH(nodoDis, nodoPart, sb, comando, tipoFormateo, cantidadLWH) {
		//ir a crear la carpeta root
		//ir a crear el archivo root
		var avd *AVDStruct
		avd = new(AVDStruct)
		crearAVD(avd, "/", nameProper) //ya tengo la raiz, ahora ir a grabarlo al inicio del avdStruct y ocupar el espacio en el bitmap

		resultBool, resultaArrBitmap := ExtrarBitmap(nodoDis.path, comando, int(sb.SbApBitMapArbolDirectorio), int(sb.SbArbolVirtualCount), "AVD")
		if resultBool {

			resultaArrBitmap[0] = 1
			if GuardarBitmap(comando, resultaArrBitmap, nodoDis.path, int(sb.SbApBitMapArbolDirectorio), "AVD") {

				if GuardarAVD(comando, avd, nodoDis.path, int(sb.SbApArbolDirectorio)) {
					Separar("mkfile -id->" + nodoPart.id + " -PatH->\"/users.txt\" -p -cont->\"1, G, root \\n1, U, root , root , 201602890\\n\"")
					PrintAviso(comando, "Excelente Eduardo pudimos guardar la raiz y el archivo users.txt, tu puedes compa :)")
					PrintAviso(comando, "La particion fue correctamente formateada y adicionalmente se creo la carpeta Raiz y el archivo users.txt")
					return
				} else {
					PrintError(ERROR, "No fue posible guardar la carpeta raiz")
					return
				}
			} else {
				PrintError(ERROR, "Error al guardar el bitmap despues de ocupar la posicion de la raiz, es decir la posicion 0")
				return
			}

		} else {
			PrintError(ERROR, "Error al extraer el bitmap para extraer la posicion de la raiz")
			return
		}

	} else {
		PrintError(ERROR, "Error al formatear la particion con el S.A LWH")
		return
	}
}

//formatearLWH formatea la particion a LWH, genera SB, guarda SB, con toda la informacion para la particion
func formatearLWH(nodoDis *NodoDisco, nodoPart *NodoParticion, sb *SuperBootStruct, comando CONSTCOMANDO, tipoFormateo string, cantidadLWH *int64) bool {

	path := nodoDis.path
	nameParticion := nodoPart.nombre
	resultFormula, partStart, tamanioPart := cantidadEstructuras(nodoPart) // != -1

	if resultFormula == -1 {
		PrintError(ERROR, "Error al calcular el numero de estructuras para la particion")
		return false
	}
	if tipoFormateo == "FULL" {
		if limpiarArchivo(comando, path, int(tamanioPart), int(partStart)) {
			PrintAviso(comando, "A la particion se le aplico el formateo FULL")
		} else {
			PrintError(comando, "No se pudo aplicar el formateo FULL a la particion")
			return false
		}
	} else {
		PrintAviso(comando, "A la particion se le aplico el formateo FAST")
	}

	//TODO: CANTIDAD DE ESTRUCTURAS
	//TODO: CANTIDAD DE ESTRUCTURAS FREES
	//TODO: CANTIDAD DE MONTADOS LWH
	//TODO: INICIOS DE BITMAPS Y LOS PRINCIPALES, LOG [TOMAR EN CUENTA EL TAMANIO DEL SB]
	//TODO: PRIMER LIBRE EN LOS BITMAPS

	sizeSB := int64(binary.Size(SuperBootStruct{}))
	sizeAVD := int64(binary.Size(AVDStruct{}))
	sizeDD := int64(binary.Size(DDStruct{}))
	sizeInodo := int64(binary.Size(InodoStruct{}))
	sizeBD := int64(binary.Size(BloqueDeDatosStruct{}))
	sizeBitacora := int64(binary.Size(BitacoraStruct{}))

	cantidadEstruAVD := resultFormula
	cantidadEstruDD := resultFormula
	cantidadEstruInodo := resultFormula * 5
	cantidadEstruBD := resultFormula * 5 * 4
	cantidadEstruLog := resultFormula

	cantidadMontajes := *cantidadLWH
	*cantidadLWH++

	apStartBitmapAVD := partStart + sizeSB
	apStartAVD := apStartBitmapAVD + cantidadEstruAVD //al inicio,....

	apStartBitmapDD := apStartAVD + (cantidadEstruAVD * sizeAVD)
	apStartDD := apStartBitmapDD + cantidadEstruDD //al inicio,....

	apStartBitmapInodo := apStartDD + (cantidadEstruDD * sizeDD)
	apStartInodo := apStartBitmapInodo + cantidadEstruInodo //al inicio,....

	apStartBitmapBD := apStartInodo + (cantidadEstruInodo * sizeInodo)
	apStartBD := apStartBitmapBD + cantidadEstruBD //al inicio,....

	apStartLog := apStartBD + (cantidadEstruBD * sizeBD) //al inicio,....

	apStartSBCopy := apStartLog + (cantidadEstruLog * sizeBitacora)

	construirSB(sb, path, cantidadEstruAVD, cantidadEstruDD, cantidadEstruInodo, cantidadEstruBD,
		cantidadEstruAVD, cantidadEstruDD, cantidadEstruInodo, cantidadEstruBD,
		cantidadMontajes,
		apStartBitmapAVD, apStartAVD,
		apStartBitmapDD, apStartDD,
		apStartBitmapInodo, apStartInodo,
		apStartBitmapBD, apStartBD,
		apStartLog,
		apStartSBCopy,
		1, 1, 1, 1)

	if GuardarSB(comando, sb, path, int(partStart), false) {

		limpiarBitmapAVD := limpiarArchivo(comando, path, int(cantidadEstruAVD), int(apStartBitmapAVD))
		limpiarBitmapDD := limpiarArchivo(comando, path, int(cantidadEstruDD), int(apStartBitmapDD))
		limpiarBitmapInodo := limpiarArchivo(comando, path, int(cantidadEstruInodo), int(apStartBitmapInodo))
		limpiarBitmapBD := limpiarArchivo(comando, path, int(cantidadEstruBD), int(apStartBitmapBD))
		limpiarBitmapLog := limpiarArchivo(comando, path, int(cantidadEstruLog), int(apStartLog))
		if limpiarBitmapAVD && limpiarBitmapDD && limpiarBitmapInodo && limpiarBitmapBD && limpiarBitmapLog {
			PrintAviso(comando, "Bitmaps inicializados correctamente [Particion: "+nameParticion+", Disco: "+path+"]")

			if GuardarSB(comando, sb, path, int(apStartSBCopy), true) {
				PrintAviso(comando, "Se escribo el SB Copia correctamente en la particion en el [Byte: "+strconv.Itoa(int(apStartSBCopy))+"]")
				PrintAviso(comando, "Particion formateada correctamente con el S.A LWH [Particion: "+nameParticion+", Disco: "+path+"]")
				nodoPart.isPartFormatLWH = true
				return true

			} else {
				PrintError(ERROR, "Error al guardar el SB copia en la particion")
				PrintAviso(comando, "Particion formateada con el S.A LWH, solamente que si recuperacion [Particion: "+nameParticion+", Disco: "+path+"]")
				nodoPart.isPartFormatLWH = true
				return true
			}

		} else {
			PrintError(ERROR, "Error al escribir el inicio de los bitmaps [AVD: "+strconv.FormatBool(limpiarBitmapAVD)+", DD: "+strconv.FormatBool(limpiarBitmapDD)+", Inodo: "+strconv.FormatBool(limpiarBitmapInodo)+", BD: "+strconv.FormatBool(limpiarBitmapBD)+", Log: "+strconv.FormatBool(limpiarBitmapLog)+"]")
			return false
		}
	} else {
		PrintError(ERROR, "No fue posible guardar el SB en la particion del disco")
		return false
	}

}

//cantidadEstructuras retorna el numero de struct's a utilizar y el tamanio del sistema relativamente [cantidadEstructuras, partStart]
func cantidadEstructuras(nodoPart *NodoParticion) (int64, int64, int64) {
	//TODO:TAMANIO TOTAL DEL SISTEMA
	//TODO:NUMERO DE ESTRUCTURAS

	tamanioParticion := 0 //int
	partStart := -1

	var partition *PartitionStruct
	partition = nodoPart.partition

	var ebr *EBRStruct
	ebr = nodoPart.ebr

	if partition != nil { //si es primaria
		tamanioParticion = int(partition.PartSize)
		partStart = int(partition.PartStart)
	} else if ebr != nil { //si es logica
		tamanioParticion = int(ebr.PartSize) - binary.Size(ebr) //esto porque incluyo el tamanio del EBR en el tamanio de la particion
		partStart = int(ebr.PartStart) + binary.Size(ebr)       //esto porque incluyo el tamanio del EBR en el tamanio de la particion
	} else { //en mkfs no se puede formatear otro tipo de particion
		PrintError(ERROR, "Error, en mkfs no se puede formatear otra particion que no sea Primario o Logica")
		return -1, -1, -1
	}

	sb := SuperBootStruct{}
	avd := AVDStruct{}
	dd := DDStruct{}
	inodo := InodoStruct{}
	bloqueDatos := BloqueDeDatosStruct{}
	bitacora := BitacoraStruct{}

	numeroEstructuras := (float64(tamanioParticion - (2 * binary.Size(sb)))) / (float64(27 + binary.Size(avd) + binary.Size(dd) + (5*binary.Size(inodo) + (20 * binary.Size(bloqueDatos)) + binary.Size(bitacora))))
	return int64(math.Floor(numeroEstructuras)), int64(partStart), int64(tamanioParticion) //el entero menor
}

//TODO: VERIFICAR SI ES NECESARIO AUMENTAR EL TAMANIO A LOS ARREGLOS DE BYTES
//TODO: ARREGLAR COMO CONSTRUYO EL STRING PARA LAS FECHAS EN LOS DEMAS LUGARES, EJEMPLO ESTA FUNCION

//construirSB construir un SB al inicio
func construirSB(sb *SuperBootStruct, path string, //al inicio,...(modificacion de agregar/quitar espacio al s.a lwh)
	cantidadEstruAVD int64, cantidadEstruDD int64, cantidadEstruInodo int64, cantidadEstruBD int64, //al inicio,...
	cantidadEstruAVDFree int64, cantidadEstruDDFree int64, cantidadEstruInodoFree int64, cantidadEstruBDFree int64, //IR modificando
	cantidadMontajes int64, //al inicio,... //TODO: UNA VARIABLE GLOBAL QUE ME DIGA CUANTOS S.A LWH VOY MONTANDO (HACIENDO)
	apStartBitmapAVD int64, apStartAVD int64, //al inicio,....
	apStartBitmapDD int64, apStartDD int64, //al inicio,....
	apStartBitmapInodo int64, apStartInodo int64, //al inicio,....
	apStartBitmapBD int64, apStartBD int64, //al inicio,....
	apStartLog int64,
	apSBCopy int64, //al inicio,....
	firstFreeBitAVD int64, firstFreeBitDD int64, firstFreeBitInodo int64, firstFreeBitBD int64) {

	nombreComodin := [50]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	sb.SbNombreHd = nombreComodin
	copy(sb.SbNombreHd[:], path)

	sb.SbArbolVirtualCount = cantidadEstruAVD     //cantidad de estructuras en el arbol virtual de DIRECTORIOS //ARBOL VIRTUAL DIRECTORIOS
	sb.SbDetalleDirectorioCount = cantidadEstruDD //cantidad de estructuras en el detalle de DIRECTORIOS //DETALLE DE DIRECTORIOS
	sb.SbInodosCount = cantidadEstruInodo         //cantidad de Inodos en la tabla de Inodos //TABLA DE INODOS
	sb.SbBloquesCount = cantidadEstruBD           //cantidad de bloques de datos //BLOQUES DE DATOS

	sb.SbArbolVirtualFree = cantidadEstruAVDFree     //cantidad de estructuras en el arbol de directorios libres // ARBOL VIRTUAL DIRECTORIOS LIBRES
	sb.SbDetalleDirectorioFree = cantidadEstruDDFree //cantidad de estructuras en el detalle de directoriso libres //DETALLE DE DIRECTORIOS LIBRES
	sb.SbInodosFree = cantidadEstruInodoFree         //cantidad de inodos en la tabla de inodos en la tabla de inodos libres //TABLA DE INODOS LIBRES
	sb.SbBloquesFree = cantidadEstruBDFree           //cantidad de bloques de datos libres //BLOQUE DE DATOS LIBRES

	fechaCreacion := [25]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	sb.SbDateCreacion = fechaCreacion //fecha de creacion del sistema
	copy(sb.SbDateCreacion[:], time.Now().Format("01-02-2006 15:04:05"))

	fechaUltimoMontaje := [25]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	sb.SbDateUltimoMontaje = fechaUltimoMontaje //ultima fecha de montaje
	copy(sb.SbDateUltimoMontaje[:], time.Now().Format("01-02-2006 15:04:05"))

	sb.SbMontajesCount = cantidadMontajes //cantidad de montajes del sistema LWH

	sb.SbApBitMapArbolDirectorio = apStartBitmapAVD //apuntador al inicio del bitmap del arbol virtual de directorio //APUNTADOR AL INICIO DEL BITMAP DEL ARBOL VIRTUAL DE DIRECTORIO
	sb.SbApArbolDirectorio = apStartAVD             //apuntador al inicio del arbol virtual de directorio

	sb.SbApBitmapDetalleDirectorio = apStartBitmapDD //apuntador al inicio del bitmap de detalle de directorio
	sb.SbApDetalleDirectorio = apStartDD             //apuntador al inicio del detalle directorio

	sb.SbApBitmapTablaInodo = apStartBitmapInodo //apuntador al inicio del bitmap de la tabla de inodos
	sb.SbApTablaInodo = apStartInodo             //apuntador al inicio de la tabla de inodos

	sb.SbApBitmapBloques = apStartBitmapBD //apuntador al inicio del bitmap de bloques de datos
	sb.SbApBloques = apStartBD             //apuntador al inicio del bloque de datos

	sb.SbApLog = apStartLog //apuntador al inicio del log o bitacora

	sb.SBApSBCopy = apSBCopy //apuntador al inicio de la copia del SB

	avd := AVDStruct{}
	dd := DDStruct{}
	inodo := InodoStruct{}
	bd := BloqueDeDatosStruct{}

	sb.SbSizeStructArbolDirectorio = int64(binary.Size(avd))  //tamanio de una estructura del arbol virtual de directorio
	sb.SbSizeStructDetalleDirectorio = int64(binary.Size(dd)) //tamanio de la estructura de un detalle de directorio
	sb.SbSizeStructInodo = int64(binary.Size(inodo))          //tamanio de la estructura de un inodo
	sb.SbSizeStructBloque = int64(binary.Size(bd))            //tamanio de la estructura de un bloque de datos

	sb.SbFirstFreeBitArbolDirectorio = firstFreeBitAVD  //primer bit libre en el bitmap arbol de directorio
	sb.SbFirstFreeBitDetalleDirectorio = firstFreeBitDD //primer bit libre en el bitmap detalle de directorio
	sb.SbFirstFreeBitTablaInodo = firstFreeBitInodo     //primer bit en el bitmap de inodo
	sb.SbFirstFreeBitBloques = firstFreeBitBD           //primer bit libre en el bitmap de bloques de datos

	sb.SbMagicNum = 201602890 //numero de carnet del estudiante

} //IR modificando

//actualizandoSB ir actualizando un SB
func actualizandoSB(sb *SuperBootStruct,
	cantidadEstruAVDFree int64, cantidadEstruDDFree int64, cantidadEstruInodoFree int64, cantidadEstruBDFree int64,
	firstFreeBitAVD int64, firstFreeBitDD int64, firstFreeBitInodo int64, firstFreeBitBD int64) {

	sb.SbArbolVirtualFree = cantidadEstruAVDFree     //cantidad de estructuras en el arbol de directorios libres // ARBOL VIRTUAL DIRECTORIOS LIBRES
	sb.SbDetalleDirectorioFree = cantidadEstruDDFree //cantidad de estructuras en el detalle de directoriso libres //DETALLE DE DIRECTORIOS LIBRES
	sb.SbInodosFree = cantidadEstruInodoFree         //cantidad de inodos en la tabla de inodos en la tabla de inodos libres //TABLA DE INODOS LIBRES
	sb.SbBloquesFree = cantidadEstruBDFree           //cantidad de bloques de datos libres //BLOQUE DE DATOS LIBRES

	sb.SbFirstFreeBitArbolDirectorio = firstFreeBitAVD  //primer bit libre en el bitmap arbol de directorio
	sb.SbFirstFreeBitDetalleDirectorio = firstFreeBitDD //primer bit libre en el bitmap detalle de directorio
	sb.SbFirstFreeBitTablaInodo = firstFreeBitInodo     //primer bit en el bitmap de inodo
	sb.SbFirstFreeBitBloques = firstFreeBitBD           //primer bit libre en el bitmap de bloques de datos

}

//TODO: REVISAR EN DONDE MAS PUEDO IMPLEMENTAR ESTO
//construirPartition construye un *PartitionStruct
func construirPartition(partitionStruct *PartitionStruct, status byte, tipo byte, fit byte, start int64, size int64, name string) {
	nombreComodin := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}

	partitionStruct.PartStatus = status
	partitionStruct.PartType = tipo
	partitionStruct.PartFit = fit
	partitionStruct.PartStart = start
	partitionStruct.PartSize = size
	partitionStruct.PartName = nombreComodin
	if name != "" {
		copy(partitionStruct.PartName[:], name)
	}
}

//ParametroTipo scanner para -tipo
func ParametroTipo(comando CONSTCOMANDO, mapa map[string]string) int {
	if val, ok := mapa["TIPO"]; ok {
		val = strings.ToUpper(val)

		if val == "FAST" || val == "FULL" {
			if val == "FAST" {
				mapa["TIPO"] = "FAST"
				return 0
			} else if val == "FULL" {
				mapa["TIPO"] = "FULL"
				return 0
			} else {
				return -1
			}
		} else {
			PrintError(comando, "El valor del parametro tipo no es correcto ["+val+"]")
			return -1
		}

	} else {
		PrintAviso(comando, "El parametro opcional tipo no esta en la sentencia...")
		PrintAviso(comando, "Se le asignara el valor predefinido [FULL] Formateo Completo")
		mapa["TIPO"] = "FULL"
		return 0
	}
}

//limpiarArchivo limpia el archivo con 0 desde un [inico-fin(tamanio)]
func limpiarArchivo(comando CONSTCOMANDO, path string, tamanioRelleno int, start int) bool {
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		PrintError(ERROR, "Error al abrir el archivo")
		log.Fatal(err)
		return false
	}

	relleno := make([]byte, tamanioRelleno)
	rellenoP := &relleno

	bufferBinario := new(bytes.Buffer)
	if binary.Write(bufferBinario, binary.BigEndian, rellenoP); err != nil {
		PrintError(ERROR, "Error al hacer la conversion a binario")
		fmt.Println("binary.Write failed:", err)
		file.Close()
		return false
	}

	if WriteBytes2(file, bufferBinario.Bytes(), start, 0, comando) {
		PrintAviso(comando, "Se limpio el disco correctamente [Desde: "+strconv.Itoa(start)+", Hasta: "+strconv.Itoa(start+tamanioRelleno)+"]")
		return true
	}
	PrintError(ERROR, "Error al limpiar en el disco [Desde: "+strconv.Itoa(start)+", Hasta: "+strconv.Itoa(start+tamanioRelleno)+"]")
	return false

}

//======================================================================================================================

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++REP

//TODO: HACERLES SUS RETURN BOOL Y QUE LO RECIBAN TAMBIEN PARA REPORTAR ERRORES

//ComandoRep ejecuta el comando Rep
func ComandoRep(nombreReporte string, nombreArchivo string, nodoDisco *NodoDisco, nodoParticion *NodoParticion, indexDisco int, indexParticion int, comando CONSTCOMANDO, resultRuta bool, ruta string) {
	switch nombreReporte {
	case "mbr":
		//resultMBR := RepMBR(nombreArchivo, nodoDisco, comando)
		RepMBR(nombreArchivo, nodoDisco, comando)
	case "disk":
		RepDISK(nombreArchivo, nodoDisco, comando)
	case "sb":
		RepSB(nombreArchivo, nodoDisco, nodoParticion, comando)
	case "bm_arbdir":
		RepBitmap(nombreArchivo, nodoDisco, nodoParticion, comando, 1) //crear un archivo con el bitmap y 20 caracteres por linea separados por |
	case "bm_detdir":
		RepBitmap(nombreArchivo, nodoDisco, nodoParticion, comando, 2)
	case "bm_inode":
		RepBitmap(nombreArchivo, nodoDisco, nodoParticion, comando, 3)
	case "bm_block":
		RepBitmap(nombreArchivo, nodoDisco, nodoParticion, comando, 4)
	case "bitacora":
	case "directorio":
		RepDirectorio(nombreArchivo, nodoDisco, nodoParticion, comando)
	case "tree_file":
		if resultRuta {
			arrDir := ArrDir(ruta)
			if arrDir != nil {
				archivo := string(arrDir[len(arrDir)-1][:])
				arrDir = arrDir[:len(arrDir)-1]
				RepTreeFile(nombreArchivo, nodoDisco, nodoParticion, arrDir, archivo, comando)
			} else {
				PrintError(ERROR, "Existe algun error con el parametro ruta")
				return
			}
		} else {
			PrintError(ERROR, "El parametro ruta no viene para el reporte por lo tanto no se puede proceder a generar el reporte")
			return
		}
	case "tree_directorio":
		if resultRuta {
			arrDir := ArrDir(ruta)
			if arrDir != nil {
				RepTreeDirectorio(nombreArchivo, nodoDisco, nodoParticion, arrDir, comando)
			} else {
				PrintError(ERROR, "Existe algun error con el parametro ruta")
				return
			}
		} else {
			PrintError(ERROR, "El parametro ruta no viene para el reporte por lo tanto no se puede proceder a generar el reporte")
			return
		}
	case "tree_complete":
		RepTreeComplete(nombreArchivo, nodoDisco, nodoParticion, comando)
	case "ls":
	default:
		PrintError(ERROR, "El nombre del reporte no es correcto [Nombre del reporte: "+nombreReporte+"]")
		return
	}
}

func RepTreeFile(pathReporte string, nodoDisco *NodoDisco, nodoParticion *NodoParticion, arrDir [][50]byte, archivo string, comando CONSTCOMANDO) {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		texto := "digraph {\ngraph[pad=\"0.5\", nodesep=\"0.5\", ranksep=\"2\"];\nnode[shape=plain];\nrankdir=LR;\n"
		_, valTexto := RepTreeFileTexto(nodoDisco.path, sb, avd, arrDir, archivo, 0, 0, comando)
		texto += valTexto + "\n"
		texto += "}\n"
		if generadorImagen(pathReporte, texto, comando) {
			PrintAviso(comando, "Imagen generada correctamente para el reporte Tree File [Particion nombre: "+nodoParticion.nombre+"]")
			return
		} else {
			PrintError(ERROR, "Error al generar el reporte para el reporte Tree File [Particion nombre: "+nodoParticion.nombre+"]")
			return
		}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return
	}
}

func RepTreeFileTexto(pathDisco string, sb *SuperBootStruct, avd *AVDStruct, arrDir [][50]byte, archivo string, contadorArr int64, posBitmapAVD int64, comando CONSTCOMANDO) (bool, string) {

	avd = new(AVDStruct)
	if ExtrarAVD(pathDisco, comando, avd, int(sb.SbApArbolDirectorio+(posBitmapAVD*sb.SbSizeStructArbolDirectorio))) {

		if int(contadorArr) < len(arrDir) {

			if avd.AVDNombreDirectorio == arrDir[contadorArr] {
				//ADENTRO DE EL AVD
				contadorArr++

				texto := "AVD_" + strconv.Itoa(int(posBitmapAVD)) + "[label=<\n"
				texto += "<table border='0' cellborder='1' cellspacing='0'>\n"
				texto += "<tr><td COLSPAN='2' BGCOLOR='#808000'><b>AVD " + strconv.Itoa(int(posBitmapAVD)) + "</b></td></tr>\n"
				texto += "<tr><td><b>AVDFechaCreacion</b></td><td>" + string(bytes.Trim(avd.AVDFechaCreacion[:], "0")) + "</td></tr>\n"
				texto += "<tr><td><b>AVDNombreDirectorio</b></td><td>" + string(bytes.Trim(avd.AVDNombreDirectorio[:], "0")) + "</td></tr>\n"
				for i, val := range avd.AVDApArraySubdirectorios {
					texto += "<tr><td><b>aptr_" + strconv.Itoa(i+1) + "</b></td><td port='" + strconv.Itoa(i) + "'>" + strconv.Itoa(int(val)) + "</td></tr>\n"
				}
				texto += "<tr><td><b>detalle D</b></td><td port='" + strconv.Itoa(6) + "'>" + strconv.Itoa(int(avd.AVDApDetalleDirectorio)) + "</td></tr>\n"
				texto += "<tr><td><b>aptr_ind</b></td><td port='" + strconv.Itoa(7) + "'>" + strconv.Itoa(int(avd.AVDApArbolVirtualDirectorio)) + "</td></tr>\n"
				texto += "<tr><td><b>AVDProper</b></td><td>" + string(bytes.Trim(avd.AVDProper[:], "0")) + "</td></tr>\n"
				texto += "</table>\n"
				texto += ">];\n"

				if int(contadorArr) != len(arrDir) { //sino

					for i, val := range avd.AVDApArraySubdirectorios {
						if val != -1 {
							if valBool, valString := RepTreeFileTexto(pathDisco, sb, avd, arrDir, archivo, contadorArr, val, comando); valBool {
								PrintAviso(comando, "Ya fue encontrado el PATH completo")
								texto += "AVD_" + strconv.Itoa(int(posBitmapAVD)) + ":" + strconv.Itoa(i) + " -> AVD_" + strconv.Itoa(int(val)) + "\n" //relacion
								texto += valString
							}
						}
					}
					if avd.AVDApArbolVirtualDirectorio != -1 { //si tiene #
						contadorArr--
						if valBool, valString := RepTreeFileTexto(pathDisco, sb, avd, arrDir, archivo, contadorArr, avd.AVDApArbolVirtualDirectorio, comando); valBool {
							PrintAviso(comando, "Ya fue encontrado el PATH completo")
							texto += "AVD_" + strconv.Itoa(int(posBitmapAVD)) + ":" + strconv.Itoa(7) + " -> AVD_" + strconv.Itoa(int(avd.AVDApArbolVirtualDirectorio)) + "\n" //relacion
							texto += valString
						}
					}

				} else { //==len()//seguir por explorar su dd
					PrintAviso(comando, "Ya fue encontrado el PATH completo")
					if avd.AVDApDetalleDirectorio != -1 {
						var dd *DDStruct
						dd = new(DDStruct)
						nombreComodin := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
						copy(nombreComodin[:], archivo)
						if valBool3, valTexto3 := RepTreeFileTextoDD(pathDisco, sb, avd.AVDApDetalleDirectorio, dd, comando, nombreComodin); valBool3 {
							textoDD := "AVD_" + strconv.Itoa(int(posBitmapAVD)) + ":" + strconv.Itoa(6) + " -> DD_" + strconv.Itoa(int(avd.AVDApDetalleDirectorio)) + "\n" //relacion
							textoDD += valTexto3 + "\n"
							texto += textoDD
						}
					}
				}

				return true, texto
				//ADENTRO DE EL AVD
			} else {
				return false, "" //false
			}

		} else {
			return false, "" //false
		}

	} else {
		PrintError(ERROR, "Error al extraer el [AVD]")
		return false, ""
	}
}

func RepTreeFileTextoDD(path string, sb *SuperBootStruct, posBitmapDD int64, dd *DDStruct, comando CONSTCOMANDO, DDFileNombre [16]byte) (bool, string) {

	dd = new(DDStruct)
	if ExtrarDD(path, comando, dd, int(sb.SbApDetalleDirectorio+(posBitmapDD*sb.SbSizeStructDetalleDirectorio))) {

		texto := "DD_" + strconv.Itoa(int(posBitmapDD)) + "[label=<\n"
		texto += "<table border='0' cellborder='1' cellspacing='0'>\n"
		texto += "<tr><td COLSPAN='2' BGCOLOR='#FFFF00'><b>DD " + strconv.Itoa(int(posBitmapDD)) + "</b></td></tr>\n"
		for i, val := range dd.DDArrayFile {
			if string(bytes.Trim(val.DDFileNombre[:], "0")) != "" {
				texto += "<tr><td><b>" + string(bytes.Trim(val.DDFileNombre[:], "0")) + "</b></td><td port='" + strconv.Itoa(i) + "'>" + strconv.Itoa(int(val.DDFileApInodo)) + "</td></tr>\n"
			} else {
				texto += "<tr><td><b>aptr_" + strconv.Itoa(i+1) + "</b></td><td port='" + strconv.Itoa(i) + "'>" + strconv.Itoa(int(val.DDFileApInodo)) + "</td></tr>\n"
			}
		}
		texto += "<tr><td><b>aptr_ind</b></td><td port='" + strconv.Itoa(5) + "'>" + strconv.Itoa(int(dd.DDApDetalleDirectorio)) + "</td></tr>\n"
		texto += "</table>\n"
		texto += ">];\n"

		bandera := false
		for i, val := range dd.DDArrayFile { //ir a traer sus inodos
			if (val.DDFileApInodo != -1) && (val.DDFileNombre == DDFileNombre) { //==este ocupado y sera ==nombre
				bandera = true
				var inodo *InodoStruct
				inodo = new(InodoStruct)
				if valBool1, valTexto1 := RepInodoTexto(path, comando, inodo, val.DDFileApInodo, sb); valBool1 {
					res := "DD_" + strconv.Itoa(int(posBitmapDD)) + ":" + strconv.Itoa(i) + " -> INODO_" + strconv.Itoa(int(val.DDFileApInodo)) + "\n"
					res += valTexto1 + "\n"
					texto += res
					//return valBool1, res
				}
			}
		}

		if dd.DDApDetalleDirectorio != -1 { //si tiene un indirecto
			if valBool1, valTexto1 := RepTreeFileTextoDD(path, sb, dd.DDApDetalleDirectorio, dd, comando, DDFileNombre); valBool1 {
				res := "DD_" + strconv.Itoa(int(posBitmapDD)) + ":" + strconv.Itoa(5) + " -> DD_" + strconv.Itoa(int(dd.DDApDetalleDirectorio)) + "\n"
				res += valTexto1 + "\n"
				texto += res
				//return valBool1, res
			}
		}

		if bandera {
			return true, texto
		} else {
			PrintAviso(comando, "No existe un archivo con ese nombre amigo")
			return false, ""
		}

	} else {
		PrintError(ERROR, "Error al extraer el dd del archivo para la carpeta")
		return false, ""
	}
}

//RepTreeDirectorio reporte desde una carpeta en especifico
func RepTreeDirectorio(pathReporte string, nodoDisco *NodoDisco, nodoParticion *NodoParticion, arrDir [][50]byte, comando CONSTCOMANDO) {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		posAvdBitmapRoot := int64(0)
		contadorArr := int64(0)
		_, _, _, valPos := MkdirRecorrer(nodoDisco.path, avd, arrDir, contadorArr, sb.SbApArbolDirectorio, posAvdBitmapRoot, sb.SbSizeStructArbolDirectorio, 0, 1, comando)

		if valPos != -1 {

			texto := "digraph {\ngraph[pad=\"0.5\", nodesep=\"0.5\", ranksep=\"2\"];\nnode[shape=plain];\nrankdir=LR;\n"
			_, valTexto := RepTreeCompleteTexto(comando, nodoDisco.path, avd, sb.SbApArbolDirectorio, valPos, sb.SbSizeStructArbolDirectorio, sb)
			texto += valTexto + "\n"
			texto += "}\n"
			if generadorImagen(pathReporte, texto, comando) {
				PrintAviso(comando, "Imagen generada correctamente para el reporte Tree Directorio [Particion nombre: "+nodoParticion.nombre+"]")
				return
			} else {
				PrintError(ERROR, "Error al generar el reporte para el reporte Tree Directorio [Particion nombre: "+nodoParticion.nombre+"]")
				return
			}

		} else {
			PrintError(ERROR, "Error al extraer la carpeta desde donde se debe de graficar, posiblemente no existe o algun otro problema")
			return
		}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return
	}
}

//RepTreeComplete reporte completo del sistema de archivos
func RepTreeComplete(pathReporte string, nodoDisco *NodoDisco, nodoParticion *NodoParticion, comando CONSTCOMANDO) {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		texto := "digraph {\ngraph[pad=\"0.5\", nodesep=\"0.5\", ranksep=\"2\"];\nnode[shape=plain];\nrankdir=LR;\n"
		_, valTexto := RepTreeCompleteTexto(comando, nodoDisco.path, avd, sb.SbApArbolDirectorio, 0, sb.SbSizeStructArbolDirectorio, sb)
		texto += valTexto + "\n"
		texto += "}\n"
		if generadorImagen(pathReporte, texto, comando) {
			PrintAviso(comando, "Imagen generada correctamente para el reporte Tree Complete[Particion nombre: "+nodoParticion.nombre+"]")
			return
		} else {
			PrintError(ERROR, "Error al generar el reporte para el reporte Tree Complete [Particion nombre: "+nodoParticion.nombre+"]")
			return
		}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return
	}
}

func RepTreeCompleteTexto(comando CONSTCOMANDO, path string, avd *AVDStruct, startAVDStructs int64, posAVDDisco int64, sizeAVDStruct int64, sb *SuperBootStruct) (bool, string) {

	avd = new(AVDStruct)
	if ExtrarAVD(path, comando, avd, int(startAVDStructs+(posAVDDisco*sizeAVDStruct))) {

		texto := "AVD_" + strconv.Itoa(int(posAVDDisco)) + "[label=<\n"
		texto += "<table border='0' cellborder='1' cellspacing='0'>\n"
		texto += "<tr><td COLSPAN='2' BGCOLOR='#808000'><b>AVD " + strconv.Itoa(int(posAVDDisco)) + "</b></td></tr>\n"
		texto += "<tr><td><b>AVDFechaCreacion</b></td><td>" + string(bytes.Trim(avd.AVDFechaCreacion[:], "0")) + "</td></tr>\n"
		texto += "<tr><td><b>AVDNombreDirectorio</b></td><td>" + string(bytes.Trim(avd.AVDNombreDirectorio[:], "0")) + "</td></tr>\n"
		for i, val := range avd.AVDApArraySubdirectorios {
			texto += "<tr><td><b>aptr_" + strconv.Itoa(i+1) + "</b></td><td port='" + strconv.Itoa(i) + "'>" + strconv.Itoa(int(val)) + "</td></tr>\n"
		}
		texto += "<tr><td><b>detalle D</b></td><td port='" + strconv.Itoa(6) + "'>" + strconv.Itoa(int(avd.AVDApDetalleDirectorio)) + "</td></tr>\n"
		texto += "<tr><td><b>aptr_ind</b></td><td port='" + strconv.Itoa(7) + "'>" + strconv.Itoa(int(avd.AVDApArbolVirtualDirectorio)) + "</td></tr>\n"
		texto += "<tr><td><b>AVDProper</b></td><td>" + string(bytes.Trim(avd.AVDProper[:], "0")) + "</td></tr>\n"
		texto += "</table>\n"
		texto += ">];\n"

		for i, val := range avd.AVDApArraySubdirectorios { //mas carpetas
			if val != -1 { //==#
				if valBool1, valTexto1 := RepTreeCompleteTexto(comando, path, avd, startAVDStructs, val, sizeAVDStruct, sb); valBool1 {
					res := valTexto1 + "\n"                                                                                             //tabla
					res += "AVD_" + strconv.Itoa(int(posAVDDisco)) + ":" + strconv.Itoa(i) + " -> AVD_" + strconv.Itoa(int(val)) + "\n" //relacion
					texto += res
					//return valBool1, res
				}
			}
		}
		if avd.AVDApArbolVirtualDirectorio != -1 { //tiene un indirecto
			if valBool2, valTexto2 := RepTreeCompleteTexto(comando, path, avd, startAVDStructs, avd.AVDApArbolVirtualDirectorio, sizeAVDStruct, sb); valBool2 {
				res2 := "AVD_" + strconv.Itoa(int(posAVDDisco)) + ":" + strconv.Itoa(7) + " -> AVD_" + strconv.Itoa(int(avd.AVDApArbolVirtualDirectorio)) + "\n" //relacion
				res2 += valTexto2 + "\n"
				//return valBool2, res2
				texto += res2
			}
		}

		if avd.AVDApDetalleDirectorio != -1 { //si tiene un DD
			var dd *DDStruct
			dd = new(DDStruct)
			if valBool3, valTexto3 := RepDDtexto(path, sb, avd.AVDApDetalleDirectorio, dd, comando); valBool3 {
				//relacion
				//tabla
				textoDD := "AVD_" + strconv.Itoa(int(posAVDDisco)) + ":" + strconv.Itoa(6) + " -> DD_" + strconv.Itoa(int(avd.AVDApDetalleDirectorio)) + "\n" //relacion
				textoDD += valTexto3 + "\n"
				texto += textoDD
			}
		}

		return true, texto

	} else {
		PrintError(ERROR, "Error al extraer el [AVD: Raiz]")
		return false, "" //una estructura vacia
	}
}

//RepTreeCompleteDetalleDtexto retorna el texto para un dd y toda su info
func RepDDtexto(path string, sb *SuperBootStruct, posBitmapDD int64, dd *DDStruct, comando CONSTCOMANDO) (bool, string) {

	dd = new(DDStruct)
	if ExtrarDD(path, comando, dd, int(sb.SbApDetalleDirectorio+(posBitmapDD*sb.SbSizeStructDetalleDirectorio))) {

		texto := "DD_" + strconv.Itoa(int(posBitmapDD)) + "[label=<\n"
		texto += "<table border='0' cellborder='1' cellspacing='0'>\n"
		texto += "<tr><td COLSPAN='2' BGCOLOR='#FFFF00'><b>DD " + strconv.Itoa(int(posBitmapDD)) + "</b></td></tr>\n"
		for i, val := range dd.DDArrayFile {
			if string(bytes.Trim(val.DDFileNombre[:], "0")) != "" {
				texto += "<tr><td><b>" + string(bytes.Trim(val.DDFileNombre[:], "0")) + "</b></td><td port='" + strconv.Itoa(i) + "'>" + strconv.Itoa(int(val.DDFileApInodo)) + "</td></tr>\n"
			} else {
				texto += "<tr><td><b>aptr_" + strconv.Itoa(i+1) + "</b></td><td port='" + strconv.Itoa(i) + "'>" + strconv.Itoa(int(val.DDFileApInodo)) + "</td></tr>\n"
			}
		}
		texto += "<tr><td><b>aptr_ind</b></td><td port='" + strconv.Itoa(5) + "'>" + strconv.Itoa(int(dd.DDApDetalleDirectorio)) + "</td></tr>\n"
		texto += "</table>\n"
		texto += ">];\n"

		for i, val := range dd.DDArrayFile { //ir a traer sus inodos
			if val.DDFileApInodo != -1 { //==#inodo
				//relacion
				//tabla
				var inodo *InodoStruct
				inodo = new(InodoStruct)
				if valBool1, valTexto1 := RepInodoTexto(path, comando, inodo, val.DDFileApInodo, sb); valBool1 {
					res := "DD_" + strconv.Itoa(int(posBitmapDD)) + ":" + strconv.Itoa(i) + " -> INODO_" + strconv.Itoa(int(val.DDFileApInodo)) + "\n"
					res += valTexto1 + "\n"
					texto += res
					//return valBool1, res
				}
			}
		}

		if dd.DDApDetalleDirectorio != -1 { //si tiene un indirecto
			if valBool1, valTexto1 := RepDDtexto(path, sb, dd.DDApDetalleDirectorio, dd, comando); valBool1 {
				res := "DD_" + strconv.Itoa(int(posBitmapDD)) + ":" + strconv.Itoa(5) + " -> DD_" + strconv.Itoa(int(dd.DDApDetalleDirectorio)) + "\n"
				res += valTexto1 + "\n"
				texto += res
				//return valBool1, res
			}
		}

		return true, texto

	} else {
		PrintError(ERROR, "Error al extraer el dd del archivo para la carpeta")
		return false, ""
	}
}

//RepInodoTexto texto para un inodo
func RepInodoTexto(path string, comando CONSTCOMANDO, inodo *InodoStruct, posBitmapInodo int64, sb *SuperBootStruct) (bool, string) {
	inodo = new(InodoStruct)
	if ExtrarInodo(path, comando, inodo, int(sb.SbApTablaInodo+(posBitmapInodo*sb.SbSizeStructInodo))) {

		texto := "INODO_" + strconv.Itoa(int(posBitmapInodo)) + "[label=<\n"
		texto += "<table border='0' cellborder='1' cellspacing='0'>\n"
		texto += "<tr><td COLSPAN='2' BGCOLOR='#45B69B'><b>Inodo " + strconv.Itoa(int(posBitmapInodo)) + "</b></td></tr>\n"
		texto += "<tr><td><b>size</b></td><td>" + strconv.Itoa(int(inodo.ISizeArchivo)) + "</td></tr>\n"
		texto += "<tr><td><b>bloques</b></td><td>" + strconv.Itoa(int(inodo.ICountBloquesAsignados)) + "</td></tr>\n"
		for i, val := range inodo.IArrayBloques {
			texto += "<tr><td><b>aptr_" + strconv.Itoa(i+1) + "</b></td><td port='" + strconv.Itoa(i) + "'>" + strconv.Itoa(int(val)) + "</td></tr>\n"
		}
		texto += "<tr><td><b>proper</b></td><td>" + string(bytes.Trim(inodo.IIdProper[:], "0")) + "</td></tr>\n"
		texto += "<tr><td><b>aptr_ind</b></td><td port='" + strconv.Itoa(4) + "'>" + strconv.Itoa(int(inodo.IApIndirecto)) + "</td></tr>\n"
		texto += "</table>\n"
		texto += ">];\n"

		for i, val := range inodo.IArrayBloques { //con sus bloques
			if val != -1 {
				var bd *BloqueDeDatosStruct
				bd = new(BloqueDeDatosStruct)
				if valBool1, valTexto1 := RepBDTexto(path, comando, bd, val, sb); valBool1 {
					res := "INODO_" + strconv.Itoa(int(posBitmapInodo)) + ":" + strconv.Itoa(i) + " -> BD_" + strconv.Itoa(int(val)) + "\n"
					res += valTexto1 + "\n"
					texto += res
					//return valBool1, res
				}
			}
		}

		if inodo.IApIndirecto != -1 { //si tiene un indirecto
			if valBool1, valTexto1 := RepInodoTexto(path, comando, inodo, inodo.IApIndirecto, sb); valBool1 {
				res := "INODO_" + strconv.Itoa(int(posBitmapInodo)) + ":" + strconv.Itoa(4) + " -> INODO_" + strconv.Itoa(int(inodo.IApIndirecto)) + "\n"
				res += valTexto1 + "\n"
				texto += res
				//return valBool1, res
			}
		}

		return true, texto

	} else {
		PrintError(ERROR, "Error al extraer el inodo del archivo")
		return false, ""
	}
}

//RepBDTexto extrae el texto para un BD
func RepBDTexto(path string, comando CONSTCOMANDO, bd *BloqueDeDatosStruct, posBitmapBD int64, sb *SuperBootStruct) (bool, string) {

	if ExtrarBD(path, comando, bd, int(sb.SbApBloques+(posBitmapBD*sb.SbSizeStructBloque))) {

		texto := "BD_" + strconv.Itoa(int(posBitmapBD)) + "[label=<\n"
		texto += "<table border='0' cellborder='1' cellspacing='0'>\n"
		texto += "<tr><td COLSPAN='2' BGCOLOR='#9A66C5'><b>Bd " + strconv.Itoa(int(posBitmapBD)) + "</b></td></tr>\n"
		texto += "<tr><td><b>contenido</b></td><td>" + string(bytes.Trim(bd.DbData[:], "0")) + "</td></tr>\n"
		texto += "</table>\n"
		texto += ">];\n"

		return true, texto

	} else {
		PrintError(ERROR, "Error al extraer un BD del archivo")
		return false, ""
	}
}

//RepDirectorio genera el reporte de Directorio
func RepDirectorio(pathReporte string, nodoDisco *NodoDisco, nodoParticion *NodoParticion, comando CONSTCOMANDO) {
	var avd *AVDStruct
	avd = new(AVDStruct)
	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)
	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {

		//TODO: INICIALIAR AVD AQUI O ADENTRO DE LA FUNCION
		texto := "digraph {\ngraph[pad=\"0.5\", nodesep=\"0.5\", ranksep=\"2\"];\nnode[shape=plain];\nrankdir=LR;\n"
		_, valTexto := RepDirectorioTexto(comando, nodoDisco.path, avd, sb.SbApArbolDirectorio, 0, sb.SbSizeStructArbolDirectorio)
		texto += valTexto + "\n"
		texto += "}\n"
		if generadorImagen(pathReporte, texto, comando) {
			PrintAviso(comando, "Imagen generada correctamente para el reporte Directorio[Particion nombre: "+nodoParticion.nombre+"]")
			return
		} else {
			PrintError(ERROR, "Error al generar el reporte para el reporte Directorio [Particion nombre: "+nodoParticion.nombre+"]")
			return
		}

	} else {
		PrintError(ERROR, "Error al extraer el sb y el sb copia del disco de la particion")
		return
	}
}

//RepDirectorio reporte de la estructura logica de todos los directorios
func RepDirectorioTexto(comando CONSTCOMANDO, path string, avd *AVDStruct, startAVDStructs int64, posAVDDisco int64, sizeAVDStruct int64) (bool, string) {

	avd = new(AVDStruct)
	if ExtrarAVD(path, comando, avd, int(startAVDStructs+(posAVDDisco*sizeAVDStruct))) {

		texto := "AVD_" + strconv.Itoa(int(posAVDDisco)) + "[label=<\n"
		texto += "<table border='0' cellborder='1' cellspacing='0'>\n"
		texto += "<tr><td COLSPAN='2' BGCOLOR='#3399ff'><b>AVD " + strconv.Itoa(int(posAVDDisco)) + "</b></td></tr>\n"
		texto += "<tr><td><b>AVDFechaCreacion</b></td><td>" + string(bytes.Trim(avd.AVDFechaCreacion[:], "0")) + "</td></tr>\n"
		texto += "<tr><td><b>AVDNombreDirectorio</b></td><td>" + string(bytes.Trim(avd.AVDNombreDirectorio[:], "0")) + "</td></tr>\n"

		for i, val := range avd.AVDApArraySubdirectorios {
			texto += "<tr><td><b>aptr_" + strconv.Itoa(i+1) + "</b></td><td port='" + strconv.Itoa(i) + "'>" + strconv.Itoa(int(val)) + "</td></tr>\n"
		}
		texto += "<tr><td><b>aptr_ind</b></td><td port='" + strconv.Itoa(6) + "'>" + strconv.Itoa(int(avd.AVDApArbolVirtualDirectorio)) + "</td></tr>\n"

		texto += "<tr><td><b>AVDProper</b></td><td>" + string(bytes.Trim(avd.AVDProper[:], "0")) + "</td></tr>\n"
		texto += "</table>\n"
		texto += ">];\n"

		for i, val := range avd.AVDApArraySubdirectorios {
			if val != -1 { //==#
				if valBool1, valTexto1 := RepDirectorioTexto(comando, path, avd, startAVDStructs, val, sizeAVDStruct); valBool1 {
					res := valTexto1 + "\n"                                                                                             //tabla
					res += "AVD_" + strconv.Itoa(int(posAVDDisco)) + ":" + strconv.Itoa(i) + " -> AVD_" + strconv.Itoa(int(val)) + "\n" //relacion
					texto += res
					//return valBool1, res
				}
			}
		}
		if avd.AVDApArbolVirtualDirectorio != -1 {
			if valBool2, valTexto2 := RepDirectorioTexto(comando, path, avd, startAVDStructs, avd.AVDApArbolVirtualDirectorio, sizeAVDStruct); valBool2 {
				res2 := "AVD_" + strconv.Itoa(int(posAVDDisco)) + ":" + strconv.Itoa(6) + " -> AVD_" + strconv.Itoa(int(avd.AVDApArbolVirtualDirectorio)) + "\n" //relacion
				res2 += valTexto2 + "\n"
				//return valBool2, res2
				texto += res2
			}
		}

		return true, texto

	} else {
		PrintError(ERROR, "Error al extraer el [AVD: Raiz]")
		return false, "" //una estructura vacia
	}
}

//++++++++++++++++++++++++++++++++++++++++++++++BITMAP AVD
//RepBitmap reporte del bitmap de cualquier estructura
func RepBitmap(pathReporte string, nodoDisco *NodoDisco, nodoParticion *NodoParticion, comando CONSTCOMANDO, bitmapTipo int) {

	var sb, sbCopia *SuperBootStruct
	sb, sbCopia = new(SuperBootStruct), new(SuperBootStruct)

	if ObtenerSBySBCopia(nodoDisco, nodoParticion, sb, sbCopia, comando) {
		start := int64(0)
		tamanio := int64(0)
		bitmap := ""

		switch bitmapTipo {
		case 1: //bm_arbdir
			start = sb.SbApBitMapArbolDirectorio
			tamanio = sb.SbArbolVirtualCount
			bitmap = "bm_arbdir"
		case 2: //bm_detdir
			start = sb.SbApBitmapDetalleDirectorio
			tamanio = sb.SbDetalleDirectorioCount
			bitmap = "bm_detdir"
		case 3: //bm_inode
			start = sb.SbApBitmapTablaInodo
			tamanio = sb.SbInodosCount
			bitmap = "bm_inode"
		case 4: //bm_block
			start = sb.SbApBitmapBloques
			tamanio = sb.SbBloquesCount
			bitmap = "bm_block"
		default:
			PrintError(ERROR, "No se reconocio el tipo de bitmap que desea generar el reporte")
			return
		}

		if resultBool, resultArr := ExtrarBitmap(nodoDisco.path, comando, int(start), int(tamanio), bitmap); resultBool {

			texto := ""
			contadorLineas := 1
			for _, val := range resultArr {
				if val == 1 {
					texto += "1"
				} else {
					texto += "0"
				}
				texto += "|"
				if contadorLineas == 20 {
					contadorLineas = 0
					texto += "\n"
				}
				contadorLineas++
			}

			if generadorImagen(pathReporte, texto, comando) {
				PrintAviso(comando, "Imagen generada correctamente del Bitmap para [Nombre: "+bitmap+"]")
				return
			} else {
				PrintError(ERROR, "Error al generar el reporte del Bitmap para [Nombre: "+bitmap+"]")
				return
			}

		} else {
			PrintError(ERROR, "Error al extraer el bitmap de [Nombre: "+bitmap+"]")
			return
		}

	} else {
		PrintError(ERROR, "Error al extraer el sb de la particion")
		return
	}
}

//++++++++++++++++++++++++++++++++++++++++++++++SB

//RepSB genera el reporte del SB
func RepSB(pathReporte string, nodoDisco *NodoDisco, nodoParticion *NodoParticion, comando CONSTCOMANDO) {

	path := nodoDisco.path

	var mbr *MBRStruct
	mbr = new(MBRStruct)

	var sb *SuperBootStruct
	sb = new(SuperBootStruct)

	if ExtrarMBR(path, comando, mbr) { //traer el MBR del disco

		isOk, startPart, _, _ := StartSizeParticion(nodoParticion)

		if isOk {
			if ExtrarSB(path, comando, sb, startPart) {
				if textoSB(path, nodoParticion.nombre, sb, pathReporte, comando) {
					PrintAviso(comando, "Reporte generado correctamente")
					return
				} else {
					PrintError(ERROR, "No fue posible generar el reporte SB")
					return
				}
			} else {
				PrintError(ERROR, "Error al extraer el SB de la particion del disco [Disco :"+path+", Particion: "+nodoParticion.nombre+"]")
				return
			}
		} else {
			PrintError(ERROR, "Error al extraer la informacion de la particion montada [Disco :"+path+", Particion: "+nodoParticion.nombre+"]")
			return
		}

	} else {
		PrintError(ERROR, "Error al extraer el MBR para generar el reporte SB [Path: "+path+"]")
		return
	}

}

//textoSB texto para del SB extraido
func textoSB(pathDisco string, namePart string, sb *SuperBootStruct, pathReporte string, comando CONSTCOMANDO) bool {

	texto := "digraph G{\nMBR [\nshape=plaintext\nlabel=<"

	texto += "<table border='0' cellborder='1' cellspacing='0' cellpadding='10'>"

	texto += "<tr>\n<td><b>Nombre</b></td>\n<td><b>Valor</b></td>\n</tr>"

	texto += "<tr>\n<td><b>sb_nombre_hd</b></td>\n<td>" + string(sb.SbNombreHd[:len(pathDisco)]) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_arbol_virtual_count</b></td>\n<td>" + strconv.Itoa(int(sb.SbArbolVirtualCount)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_detalle_directorio_count</b></td>\n<td>" + strconv.Itoa(int(sb.SbDetalleDirectorioCount)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_inodos_count</b></td>\n<td>" + strconv.Itoa(int(sb.SbInodosCount)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_bloques_count</b></td>\n<td>" + strconv.Itoa(int(sb.SbBloquesCount)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_arbol_virtual_free</b></td>\n<td>" + strconv.Itoa(int(sb.SbArbolVirtualFree)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_detalle_directorio_free</b></td>\n<td>" + strconv.Itoa(int(sb.SbDetalleDirectorioFree)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_inodos_free</b></td>\n<td>" + strconv.Itoa(int(sb.SbInodosFree)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_bloques_free</b></td>\n<td>" + strconv.Itoa(int(sb.SbBloquesFree)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_date_creacion</b></td>\n<td>" + string(sb.SbDateCreacion[:19]) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_date_ultimo_montaje</b></td>\n<td>" + string(sb.SbDateUltimoMontaje[:19]) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_montajes_count</b></td>\n<td>" + strconv.Itoa(int(sb.SbMontajesCount)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_ap_bitmap_arbol_directorio</b></td>\n<td>" + strconv.Itoa(int(sb.SbApBitMapArbolDirectorio)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_ap_arbol_directorio</b></td>\n<td>" + strconv.Itoa(int(sb.SbApArbolDirectorio)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_ap_bitmap_detalle_directorio</b></td>\n<td>" + strconv.Itoa(int(sb.SbApBitmapDetalleDirectorio)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_ap_detalle_directorio</b></td>\n<td>" + strconv.Itoa(int(sb.SbApDetalleDirectorio)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_ap_bitmap_tabla_inodo</b></td>\n<td>" + strconv.Itoa(int(sb.SbApBitmapTablaInodo)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_ap_tabla_inodo</b></td>\n<td>" + strconv.Itoa(int(sb.SbApTablaInodo)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_ap_bitmap_bloques</b></td>\n<td>" + strconv.Itoa(int(sb.SbApBitmapBloques)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_ap_bloques</b></td>\n<td>" + strconv.Itoa(int(sb.SbApBloques)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_ap_log</b></td>\n<td>" + strconv.Itoa(int(sb.SbApLog)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_ap_SBCopy</b></td>\n<td>" + strconv.Itoa(int(sb.SBApSBCopy)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_size_struct_arbol_directorio</b></td>\n<td>" + strconv.Itoa(int(sb.SbSizeStructArbolDirectorio)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_size_struct_detalle_directorio</b></td>\n<td>" + strconv.Itoa(int(sb.SbSizeStructDetalleDirectorio)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_size_struct_inodo</b></td>\n<td>" + strconv.Itoa(int(sb.SbSizeStructInodo)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_size_struct_bloque</b></td>\n<td>" + strconv.Itoa(int(sb.SbSizeStructBloque)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_first_free_bit_arbol_directorio</b></td>\n<td>" + strconv.Itoa(int(sb.SbFirstFreeBitArbolDirectorio)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_first_free_bit_detalle_directorio</b></td>\n<td>" + strconv.Itoa(int(sb.SbFirstFreeBitDetalleDirectorio)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_first_free_bit_tabla_inodo</b></td>\n<td>" + strconv.Itoa(int(sb.SbFirstFreeBitTablaInodo)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_first_free_bit_bloques</b></td>\n<td>" + strconv.Itoa(int(sb.SbFirstFreeBitBloques)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>sb_magic_num</b></td>\n<td>" + strconv.Itoa(int(sb.SbMagicNum)) + "</td>\n</tr>"

	texto += "</table>\n>\n];"
	texto += "\n}"

	if generadorImagen(pathReporte, texto, comando) {
		PrintAviso(comando, "Imagen generada correctamente del SB para [Particion: "+namePart+", Disco: "+pathDisco+"]")
		return true
	} else {
		PrintError(ERROR, "Error al generar el reporte del SB para [Particion: "+namePart+", Disco: "+pathDisco+"]")
		return false
	}
}

//++++++++++++++++++++++++++++++++++++++++++++++DISK
//TODO: FALTA ARREGLAR UNOS DETALLES DE ESTOS REPORTES

//RepDISK genera el reporte del MBR
func RepDISK(pathReporte string, nodoDisco *NodoDisco, comando CONSTCOMANDO) {

	path := nodoDisco.path

	var mbr *MBRStruct
	mbr = new(MBRStruct)

	if ExtrarMBR(path, comando, mbr) {

		resultBoolText, resultText := getTextAllMBRandEBRDISK(mbr, path, comando)

		if resultBoolText {
			PrintAviso(comando, "Todo el texto del .dot generado correctamente")
		}

		if generadorImagen(pathReporte, resultText, comando) {
			PrintAviso(comando, "Imagen generada correctamente")
		} else {
			PrintError(ERROR, "Error al generar el reporte")
			return
		}

	} else {
		PrintError(ERROR, "Error al extraer el MBR para generar el reporte DISK [Path: "+path+"]")
		return
	}

}

//getTextAllMBRandEBRDISK obtiene el texto para el reporte MBR con todo y sus EBR's si tuviera una extendida
func getTextAllMBRandEBRDISK(mbr *MBRStruct, path string, comando CONSTCOMANDO) (bool, string) {

	texto := "digraph {\nNodo [\nshape=plaintext\nlabel=<"
	texto += "<table border='0' cellborder='1' cellspacing='0' cellpadding='10'>"
	texto += "<tr>"

	texto += "<td border=\"1\" height=\"80\" width=\"30\" bgcolor=\"" + coloresDISK[1] + "\">"
	texto += "MBR<br/>Tamanio" + strconv.Itoa(binary.Size(mbr)) + "<br/>"
	texto += "Tamanio Disco" + strconv.Itoa(int(mbr.MbrTamanio)) + "<br/>"
	texto += "</td>"

	start := int64(binary.Size(mbr)) //posArchivo
	end := int64(0)

	listaLibres := make([]Libre, 0)
	ultimoEnd := int64(0)

	for i := range mbr.Partition {
		//TODO PREGUNTAR SOLO POR LAS DE STATUS ACTIVADAS DESPUES DE ELIMINAR? VER ?

		if mbr.Partition[i].PartStart != -1 { //.PartStatus == 49 (1)(si ocupado) || == 48 (0)(no ocupado)
			end = mbr.Partition[i].PartStart
			if (end - start) > 0 { //hay espacio libre
				libre := Libre{
					Lstart: int(start),
					Lend:   int(end),
				}
				listaLibres = append(listaLibres, libre)
				texto += "<td border=\"1\" height=\"80\" width=\"30\" bgcolor=\"" + coloresDISK[5] + "\">"
				texto += "Libre" + "<br/>"
				texto += "Tamanio " + strconv.Itoa(int(libre.Lend-libre.Lstart)) + "<br/>"
				texto += "Byte Inicio " + strconv.Itoa(int(libre.Lstart)) + "<br/>"
				texto += "Byte Final " + strconv.Itoa(int(libre.Lend)) + "<br/>"
				texto += "</td>"
			}
			start = mbr.Partition[i].PartStart + mbr.Partition[i].PartSize

			ultimoEnd = mbr.Partition[i].PartStart + mbr.Partition[i].PartSize
		}

		if mbr.Partition[i].PartType == 'P' { //SI ES UNA PRIMARIA

			texto += "<td border=\"1\" height=\"80\" width=\"30\" bgcolor=\"" + coloresDISK[i+1] + "\">"
			texto += "Primaria" + "<br/>"
			texto += string(bytes.Trim(mbr.Partition[i].PartName[:], "0")) + "<br/>"
			//texto += "Particion " + strconv.Itoa(i+1) + "<br/>"
			texto += "Tamanio " + strconv.Itoa(int(mbr.Partition[i].PartSize)) + "<br/>"
			texto += "Byte Inicio " + strconv.Itoa(int(mbr.Partition[i].PartStart)) + "<br/>"
			texto += "Byte Final " + strconv.Itoa(int(mbr.Partition[i].PartStart+mbr.Partition[i].PartSize)) + "<br/>"
			texto += "</td>"

		} else if mbr.Partition[i].PartType == 'E' {

			texto += "<td border=\"1\" height=\"80\" width=\"30\" bgcolor=\"" + coloresDISK[i+1] + "\">"
			texto += "<table border='0' cellborder='1' cellspacing='0' cellpadding='10'>"
			texto += "<tr>"
			texto += "<td colspan=\"100\">"

			texto += "Extendia" + "<br/>"
			texto += string(bytes.Trim(mbr.Partition[i].PartName[:], "0")) + "<br/>"
			//texto += "Particion " + strconv.Itoa(i+1) + "<br/>"
			texto += "Tamanio " + strconv.Itoa(int(mbr.Partition[i].PartSize)) + "<br/>"
			texto += "Byte Inicio " + strconv.Itoa(int(mbr.Partition[i].PartStart)) + "<br/>"
			texto += "Byte Final " + strconv.Itoa(int(mbr.Partition[i].PartStart+mbr.Partition[i].PartSize)) + "<br/>"

			texto += "</td>"
			texto += "</tr>"

			texto += "<tr>"

			resultBoolTextEBR, resultTextEBR := getTextAllEBRDISK(mbr, path, comando, mbr.Partition[i].PartStart+mbr.Partition[i].PartSize)
			if !resultBoolTextEBR {
				PrintError(ERROR, "Existio un error para generar el texto del o los EBR para el reporte igual se procede") //TODO: COLOCAR EL MISMO MSN EN EL MBR
			}
			texto += resultTextEBR
			texto += "</tr>"

			texto += "</table>"
			texto += "</td>"

		} else {
			PrintAviso(comando, "Se encontro una particion que no es de tipo primaria ni extendida ni logica, revisar eso, igual se sigue procediendo")
		}

	}

	if ultimoEnd < mbr.MbrTamanio {
		libre := Libre{
			Lstart: int(ultimoEnd),
			Lend:   int(mbr.MbrTamanio),
		}
		listaLibres = append(listaLibres, libre)
		texto += "<td border=\"1\" height=\"80\" width=\"30\" bgcolor=\"" + coloresDISK[5] + "\">"
		texto += "Libre" + "<br/>"
		texto += "Tamanio " + strconv.Itoa(int(libre.Lend-libre.Lstart)) + "<br/>"
		texto += "Byte Inicio " + strconv.Itoa(int(libre.Lstart)) + "<br/>"
		texto += "Byte Final " + strconv.Itoa(int(libre.Lend)) + "<br/>"
		texto += "</td>"
	}

	texto += "</tr>"
	texto += "</table>\n>\n];\n}"

	return true, texto
}

//getTextAllEBRDISK obtiene el texto de todos los EBR's
func getTextAllEBRDISK(mbr *MBRStruct, path string, comando CONSTCOMANDO, byteFinalPosExtendida int64) (bool, string) {

	var prevEBR *EBRStruct
	prevEBR = new(EBRStruct)
	prevEBR = getFirstEBR(mbr, path, comando) //puede ser nil//==First EBR

	if prevEBR != nil {

		textoEBR := ""
		textoEBR += "<td border=\"1\" height=\"80\" width=\"30\" bgcolor=\"" + coloresDISK[5] + "\">"
		textoEBR += "EBR<br/>" + strconv.Itoa(binary.Size(prevEBR)) + " bytes"
		textoEBR += "</td>"

		if prevEBR.PartSize > 0 {
			textoEBR += "<td border=\"1\" height=\"80\" width=\"30\" bgcolor=\"" + coloresDISK[6] + "\">"
			textoEBR += "Logica<br/>"
			textoEBR += string(bytes.Trim(prevEBR.PartName[:], "0")) + "<br/>"
			textoEBR += "Tamanio " + strconv.Itoa(int(prevEBR.PartSize)) + "<br/>"
			textoEBR += "Byte Inicio " + strconv.Itoa(int(prevEBR.PartStart)) + "<br/>"
			textoEBR += "Byte Final " + strconv.Itoa(int(prevEBR.PartStart+prevEBR.PartSize)) + "<br/>"
			textoEBR += "Byte Next " + strconv.Itoa(int(prevEBR.PartNext)) + "<br/>"
			textoEBR += "</td>"
		}

		start := prevEBR.PartStart //inicio de la extendida, aqui no puedo iniciar despues de su ebr porque no se comporta igual
		end := int64(0)

		ultimoEnd := prevEBR.PartStart + prevEBR.PartSize

		var nextEBR *EBRStruct //==nil
		nextEBR = new(EBRStruct)

		for {
			if nextEBR = getNextEBR(path, prevEBR, comando); nextEBR != nil {

				textoEBR += "<td border=\"1\" height=\"80\" width=\"30\" bgcolor=\"" + coloresDISK[5] + "\">"
				textoEBR += "EBR<br/>" + strconv.Itoa(binary.Size(nextEBR)) + " bytes"
				textoEBR += "</td>"

				textoEBR += "<td border=\"1\" height=\"80\" width=\"30\" bgcolor=\"" + coloresDISK[6] + "\">"
				textoEBR += "Logica<br/>"
				textoEBR += string(bytes.Trim(nextEBR.PartName[:], "0")) + "<br/>"
				textoEBR += "Tamanio " + strconv.Itoa(int(nextEBR.PartSize)) + "<br/>"
				textoEBR += "Byte Inicio " + strconv.Itoa(int(nextEBR.PartStart)) + "<br/>"
				textoEBR += "Byte Final " + strconv.Itoa(int(nextEBR.PartStart+nextEBR.PartSize)) + "<br/>"
				textoEBR += "Byte Next " + strconv.Itoa(int(nextEBR.PartNext)) + "<br/>"
				textoEBR += "</td>"

				end = nextEBR.PartStart
				if (end - start) > 0 { //hay espacio libre
					libre := Libre{
						Lstart: int(start),
						Lend:   int(end),
					}
					textoEBR += "<td border=\"1\" height=\"80\" width=\"30\" bgcolor=\"" + coloresDISK[5] + "\">"
					textoEBR += "Libre" + "<br/>"
					textoEBR += "Tamanio " + strconv.Itoa(int(libre.Lstart+libre.Lend)) + "<br/>"
					textoEBR += "Byte Inicio " + strconv.Itoa(int(libre.Lstart)) + "<br/>"
					textoEBR += "Byte Final " + strconv.Itoa(int(libre.Lend)) + "<br/>"
					textoEBR += "</td>"
				}
				start = nextEBR.PartStart + nextEBR.PartSize

				ultimoEnd = nextEBR.PartStart + nextEBR.PartSize

				prevEBR = nextEBR

			} else {
				PrintAviso(comando, "Se extrajo toda la informacion de los EBR hasta llegar a un nil para el reporte")
				break
			}
		}

		if ultimoEnd < byteFinalPosExtendida {
			libre := Libre{
				Lstart: int(ultimoEnd),
				Lend:   int(byteFinalPosExtendida),
			}
			textoEBR += "<td border=\"1\" height=\"80\" width=\"30\" bgcolor=\"" + coloresDISK[5] + "\">"
			textoEBR += "Libre" + "<br/>"
			textoEBR += "Tamanio " + strconv.Itoa(int(libre.Lend-libre.Lstart)) + "<br/>"
			textoEBR += "Byte Inicio " + strconv.Itoa(int(libre.Lstart)) + "<br/>"
			textoEBR += "Byte Final " + strconv.Itoa(int(libre.Lend)) + "<br/>"
			textoEBR += "</td>"
		}

		return true, textoEBR

	}
	PrintError(ERROR, "Problemas al extraer el primer EBR del disco para extraer la informacion de el")
	return false, ""

}

//+++++++++++++++++++++++++++++++++++++++++++++++MBR

//RepMBR genera el reporte del MBR
func RepMBR(pathReporte string, nodoDisco *NodoDisco, comando CONSTCOMANDO) {

	path := nodoDisco.path

	var mbr *MBRStruct
	mbr = new(MBRStruct)

	if ExtrarMBR(path, comando, mbr) {

		resultBoolText, resultText := getTextAllMBRandEBR(mbr, path, comando)

		if resultBoolText {
			PrintAviso(comando, "Todo el texto del .dot generado correctamente")
		}

		if generadorImagen(pathReporte, resultText, comando) {
			PrintAviso(comando, "Imagen generada correctamente")
		} else {
			PrintError(ERROR, "Error al generar el reporte")
			return
		}

	} else {
		PrintError(ERROR, "Error al extraer el MBR para generar el reporte MBR [Path: "+path+"]")
		return
	}

}

//getTextAllMBRandEBR obtiene el texto para el reporte MBR con todo y sus EBR's si tuviera una extendida
func getTextAllMBRandEBR(mbr *MBRStruct, path string, comando CONSTCOMANDO) (bool, string) {

	textoEBR := ""
	resultBoolean := true

	texto := "digraph G{\nMBR [\nshape=plaintext\nlabel=<"

	texto += "<table border='0' cellborder='1' cellspacing='0' cellpadding='10'>"

	texto += "<tr>\n<td><b>Nombre</b></td>\n<td><b>Valor</b></td>\n</tr>"

	texto += "<tr>\n<td><b>mbr_tamaño</b></td>\n<td>" + strconv.Itoa(int(mbr.MbrTamanio)) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>mbr_fecha_creacion</b></td>\n<td>" + string(mbr.MbrFechaCreacion[:19]) + "</td>\n</tr>"
	texto += "<tr>\n<td><b>mbr_disk_signature</b></td>\n<td>" + strconv.Itoa(int(mbr.MbrDiskSignature)) + "</td>\n</tr>"

	//TODO: SABER SI ES PRIMARIA O EXTENDIDA O LOGICA..... PERO ESTO PARA EL REPORTE DEL DISCO PORQUE TENGO QUE GENERAR DIFERENTE FIGURITA
	for i := range mbr.Partition { //POR CADA PARTICION GENERAR SU TEXTO
		//NO IMPORTA SI ES P O E

		texto += "<tr>\n<td><b>part_status_" + strconv.Itoa(i) + "</b></td>\n<td>" + string(mbr.Partition[i].PartStatus) + "</td>\n</tr>"
		texto += "<tr>\n<td><b>part_type_" + strconv.Itoa(i) + "</b></td>\n<td>" + string(mbr.Partition[i].PartType) + "</td>\n</tr>"
		texto += "<tr>\n<td><b>part_fit_" + strconv.Itoa(i) + "</b></td>\n<td>" + string(mbr.Partition[i].PartFit) + "</td>\n</tr>"
		texto += "<tr>\n<td><b>part_start_" + strconv.Itoa(i) + "</b></td>\n<td>" + strconv.Itoa(int(mbr.Partition[i].PartStart)) + "</td>\n</tr>"
		texto += "<tr>\n<td><b>part_size_" + strconv.Itoa(i) + "</b></td>\n<td>" + strconv.Itoa(int(mbr.Partition[i].PartSize)) + "</td>\n</tr>"
		texto += "<tr>\n<td><b>part_name_" + strconv.Itoa(i) + "</b></td>\n<td>" + string(bytes.Trim(mbr.Partition[i].PartName[:], "0")) + "</td>\n</tr>"

		//VER SI NO TENGO QUE AGREGAR UN SALTO AL FINAL DE CADA LINEA //LA RESPUESTA ES NO
		if mbr.Partition[i].PartType == 'E' { //SI ES UNA EXTENDIDA

			//IR A TRAER UN EBR Y ASI SUCESIVAMENTE TODOS LOS QUE TENGA LA EXTENDIDA
			resultBoolTextEBR, resultTextEBR := getTextAllEBR(mbr, path, comando)
			if !resultBoolTextEBR {
				PrintError(ERROR, "Existio un error para generar el texto del o los EBR para el reporte")
				resultBoolean = false
			}
			textoEBR += resultTextEBR

		}
	}
	texto += "</table>\n>\n];"

	texto += textoEBR

	texto += "\n}"

	return resultBoolean, texto
}

//getTextAllEBR obtiene el texto de todos los EBR's
func getTextAllEBR(mbr *MBRStruct, path string, comando CONSTCOMANDO) (bool, string) {

	var prevEBR *EBRStruct
	prevEBR = new(EBRStruct)
	prevEBR = getFirstEBR(mbr, path, comando) //puede ser nil//==First EBR

	if prevEBR != nil {

		textoEBR := ""
		contadorEBR := 1

		textoEBR += "EBR_" + strconv.Itoa(contadorEBR) + "[\nshape=plaintext\nlabel=<"
		textoEBR += "<table border='0' cellborder='1' cellspacing='0' cellpadding='10'>"
		textoEBR += "<tr>\n<td><b>Nombre</b></td>\n<td><b>Valor</b></td>\n</tr>"

		textoEBR += "<tr>\n<td><b>part_status_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + string(prevEBR.PartStatus) + "</td>\n</tr>"
		textoEBR += "<tr>\n<td><b>part_fit_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + string(prevEBR.PartFit) + "</td>\n</tr>"
		textoEBR += "<tr>\n<td><b>part_start_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + strconv.Itoa(int(prevEBR.PartStart)) + "</td>\n</tr>"
		textoEBR += "<tr>\n<td><b>part_size_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + strconv.Itoa(int(prevEBR.PartSize)) + "</td>\n</tr>"
		textoEBR += "<tr>\n<td><b>part_next_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + strconv.Itoa(int(prevEBR.PartNext)) + "</td>\n</tr>"
		textoEBR += "<tr>\n<td><b>part_name_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + string(bytes.Trim(prevEBR.PartName[:], "0")) + "</td>\n</tr>"

		textoEBR += "</table>"
		textoEBR += ">\n];"

		contadorEBR++

		var nextEBR *EBRStruct //==nil
		nextEBR = new(EBRStruct)

		for {
			if nextEBR = getNextEBR(path, prevEBR, comando); nextEBR != nil {

				//TODO: GENERAR EL TEXTO CON LA INFORMACION
				textoEBR += "EBR_" + strconv.Itoa(contadorEBR) + "[\nshape=plaintext\nlabel=<"
				textoEBR += "<table border='0' cellborder='1' cellspacing='0' cellpadding='10'>"
				textoEBR += "<tr>\n<td><b>Nombre</b></td>\n<td><b>Valor</b></td>\n</tr>"

				textoEBR += "<tr>\n<td><b>part_status_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + string(nextEBR.PartStatus) + "</td>\n</tr>"
				textoEBR += "<tr>\n<td><b>part_fit_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + string(nextEBR.PartFit) + "</td>\n</tr>"
				textoEBR += "<tr>\n<td><b>part_start_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + strconv.Itoa(int(nextEBR.PartStart)) + "</td>\n</tr>"
				textoEBR += "<tr>\n<td><b>part_size_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + strconv.Itoa(int(nextEBR.PartSize)) + "</td>\n</tr>"
				textoEBR += "<tr>\n<td><b>part_next_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + strconv.Itoa(int(nextEBR.PartNext)) + "</td>\n</tr>"
				textoEBR += "<tr>\n<td><b>part_name_" + strconv.Itoa(contadorEBR) + "</b></td>\n<td>" + string(bytes.Trim(nextEBR.PartName[:], "0")) + "</td>\n</tr>"

				textoEBR += "</table>"
				textoEBR += ">\n];"

				contadorEBR++

				prevEBR = nextEBR

			} else {
				PrintAviso(comando, "Se extrajo toda la informacion de los EBR hasta llegar a un nil para el reporte")
				break
			}
		}

		return true, textoEBR

	}
	PrintError(ERROR, "Problemas al extraer el primer EBR del disco para extraer la informacion de el")
	return false, ""

}

//ParametroNombreRep scanner nombre del comando Rep lo devulve en minuscula
func ParametroNombreRep(comando CONSTCOMANDO, mapa map[string]string) (bool, string) {

	if val, ok := mapa["NAME"]; ok {
		if val != "" {

			val = strings.ToLower(val) //minuscula
			return true, val

		}
		PrintError(ERROR, "Error el nombre del reporte viene vacio y es obligatorio")
		return false, ""
	}
	PrintError(ERROR, "El parametro obligatorio nombre no se encuentra en la sentencia")
	return false, ""
}

//ParametroIDRep ebalua el parametro ID
func ParametroIDRep(comando CONSTCOMANDO, mapa map[string]string, mountList *[]*NodoDisco) (*NodoDisco, int, *NodoParticion, int) {

	if val, ok := mapa["ID"]; ok {
		val = strings.ToLower(val) //vda1

		if strings.HasPrefix(val, "vd") {

			cortado := strings.Trim(val, "vd") //a1
			arreglo := strings.Split(cortado, "")
			return getDiscoParticionRep(mountList, arreglo[0], val, comando) //TODO: VERIFICAR QUE ESTA SINTAXIS ME FUNCIONE

		}
		PrintError(comando, "Este id no tiene la sintaxis correcta [Id: "+val+"]")
		return nil, -1, nil, -1

	}
	PrintError(ERROR, "El parametro obligatorio Id no se encuentra en la sentencia")
	return nil, -1, nil, -1

}

//getDiscoParticionRep obtiene el disco y la particion montada ya validad para el comando REP
func getDiscoParticionRep(mountList *[]*NodoDisco, letraDisco string, id string, comando CONSTCOMANDO) (*NodoDisco, int, *NodoParticion, int) {

	var nodoDis *NodoDisco
	var indexDisc = -1

	var nodoPart *NodoParticion
	var indexPart = -1

	nodoDis, indexDisc, nodoPart, indexPart = getDiscoAndParticion(mountList, letraDisco, id)

	if nodoDis != nil {

		if nodoPart != nil {
			return nodoDis, indexDisc, nodoPart, indexPart
		}
		PrintError(ERROR, "Error la particion con el id [ID: "+id+"] no esta montada dentro del disco con la letra [Letra: "+letraDisco+"]")
		return nil, -1, nil, -1

	}
	PrintError(ERROR, "Error el disco identificado con la letra unica no esta montado [Letra: "+letraDisco+"]")
	return nil, -1, nil, -1

}

//generadorImagen genera la imagen
func generadorImagen(pathReporte string, resultText string, comando CONSTCOMANDO) bool {

	PrintAviso(comando, "Generando el reporte con la informacion obtenida correctamente")

	dir, file := filepath.Split(pathReporte)
	arrFile := strings.Split(file, ".")

	err := ioutil.WriteFile(dir+arrFile[0]+".dot", []byte(resultText), 0666) //creo el archivo.dot
	if err != nil {
		PrintError(ERROR, "Error al generar el archivo .dot")
		log.Fatal(err)
		return false
	}

	pathDotFile := dir + arrFile[0] + ".dot"
	pathImageFile := pathReporte
	tps := "-T" + arrFile[1]

	//TODO: SI ME DA TIEMPO VER AQUI UN POSIBLE err PARA NOTIFICAR
	//-Tpng //-Tps //Tjpg
	cmd := exec.Command("dot", tps, pathDotFile, "-o", pathImageFile)
	cmd.Output()
	return true
}

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++UNMOUNT

//ComandoUnmount ejecuta el comando Unmount
func ComandoUnmount(comando CONSTCOMANDO, mapa map[string]string, mountList *[]*NodoDisco) {

	for k, val := range mapa {
		if k != "INSTRUCCION" && k != "UNMOUNT" { //id1

			val = strings.ToLower(val) //vda1

			if strings.HasPrefix(val, "vd") {
				cortado := strings.Trim(val, "vd") //a1
				arreglo := strings.Split(cortado, "")
				desmontandoParticion(mountList, arreglo[0], val, comando)
			} else {
				PrintError(comando, "Este id no tiene la sintaxis correcta [Id: "+val+"]")
			}
		}
	}
}

//desmontandoParticion ejecuta el comando Unmount para un id
func desmontandoParticion(mountList *[]*NodoDisco, letraDisco string, id string, comando CONSTCOMANDO) {

	var nodoDis *NodoDisco
	var indexDisc = -1

	var nodoPart *NodoParticion
	var indexPart = -1

	nodoDis, indexDisc, nodoPart, indexPart = getDiscoAndParticion(mountList, letraDisco, id)

	if nodoDis != nil {

		if nodoPart != nil {

			countPart := getContadorPartDisco(mountList, nodoDis.path)

			if countPart == 1 { //es la ultima particion montada de la lista
				PrintAviso(comando, "Se borrara la ultima particion montada del disco [Disco: "+nodoDis.path+"]")
				if !grabarDesmontado(nodoDis, nodoPart, comando) {
					PrintError(ERROR, "Error al momento de grabar la particion en el disco no se puede desmontar")
					return
				}
				deleteParticion(nodoDis.listadoParticion, indexPart) //TODO: PUEDE SER QUE LO EJECUTE O NO PORQUE IGUAL BORRARE TODO EL NODO DEL DISCO
				deleteDisco(mountList, indexDisc)
				PrintAviso(comando, "Se borro exitosamente la particion [Id: "+id+"]")
				return
			}
			//hay mas de una particion montada
			if !grabarDesmontado(nodoDis, nodoPart, comando) {
				PrintError(ERROR, "Error al momento de grabar la particion en el disco no se puede desmontar")
				return
			}
			deleteParticion(nodoDis.listadoParticion, indexPart)
			PrintAviso(comando, "Se borro exitosamente la particion [Id: "+id+"]")
			return

		}
		PrintError(ERROR, "Error la particion con el id [ID: "+id+"] no existe dentro del disco con la letra [Letra: "+letraDisco+"]")
		return

	}
	PrintError(ERROR, "Error el disco identificado con la letra unica no esta montado [Letra: "+letraDisco+"]")
	return

}

//TODO: AL DESMONTAR EXTRAER *NodoParticion y/o *NodoDisco
//TODO: EXTRAER LOS APUNTADORES DE LAS PARTICIONES Y AHI TENGO TODA LA INFORMACION PARA IR A GRABARLOS (P|L)

//grabarDesmontado graba la particion desmontada en el disco [P|L]
func grabarDesmontado(nodoDis *NodoDisco, nodoPart *NodoParticion, comando CONSTCOMANDO) bool {

	path := nodoDis.path
	//DISCO

	var mbr *MBRStruct
	mbr = new(MBRStruct)

	var partition *PartitionStruct
	partition = nodoPart.partition
	//EXTRAER EL MBR
	//BUSCAR LA PARTICION EN EL MBR
	//IGUALARLA A ESTA NUEVA
	//VALIDACIONES NECESARIAS
	//IR A GRABAR AL DISCO

	var ebr *EBRStruct
	ebr = nodoPart.ebr
	//VERIFCAR SINO TENGO QUE HACER UNAS VALIDACIONES ANTES PARA QUE TODO ESTE BIEN
	//IR A ESCRIBIRLO CON SU INICIO DIRECTAMENTE AL ARCHIVO

	if ExtrarMBR(path, comando, mbr) {
		//TODO: SABER SI ES PRIMARIA O LOGICA
		if partition != nil { //es primaria
			if AgregarParticionAlMBRUnmount(mbr, partition) { //a la estructura

				if GuardarMBR(comando, mbr, path) { //en el disco
					PrintAviso(comando, "Particion primaria desmontada correctamente y reescrita en su MBR correctamente")
					//fmt.Println(mbr)
					return true
				}
				PrintError(ERROR, "Error al guardar el MBR despues de agregarle la particion primaria desmontada en el disco [Path: "+path+"]")
				return false

			}
			PrintError(ERROR, "Error al agregar la particion primaria al MBR del disco [Path: "+path+"]")
			return false

		} else if ebr != nil {

			if GuardarEBR(comando, ebr, path, int(ebr.PartStart)) { //en el disco
				PrintAviso(comando, "Particion logica desmontada correctamente y reescrita correctamente en el disco [Path: "+path+"]")
				//fmt.Println(ebr)
				return true
			}
			PrintError(ERROR, "Error al guardar el EBR de la particion logica desmontada en el disco [Path: "+path+"]")
			return false

		} else {
			PrintError(ERROR, "Error al identificar el tipo de particion desmontada para grabarla en el disco [Path: "+path+"]")
			return false
		}
	} else {
		PrintError(ERROR, "Error al extraer el MBR para grabar la particion desmontada en el disco [Path: "+path+"]")
		return false
	}

}

//AgregarParticionAlMBRUnmount agregar una particion al MBR
func AgregarParticionAlMBRUnmount(mbr *MBRStruct, particion *PartitionStruct) bool {
	for i := 0; i < 4; i++ {
		if mbr.Partition[i].PartStart == particion.PartStart {
			mbr.Partition[i] = *particion //TODO: VERIFICAR QUE SI LE ESTOY PASANDO LOS DATOS CORRECTOS*
			return true
		}
	}
	return false
}

//getDiscoAndParticion si existe una particion con ese Id y obtengo el *NodoParticion y su pos
func getDiscoAndParticion(mountList *[]*NodoDisco, letraDisco string, id string) (*NodoDisco, int, *NodoParticion, int) {

	var valueDisco *NodoDisco //nil
	var indexDisc = -1

	var valuePart *NodoParticion //nil
	var indexPart = -1

	for indexDisc, valueDisco = range *mountList {
		if valueDisco.letraDisco == letraDisco { //DISCO CON LA MISMA LETRA
			for indexPart, valuePart = range *valueDisco.listadoParticion {
				if valuePart.id == id { //PARTICION CON EL MISMO ID
					return valueDisco, indexDisc, valuePart, indexPart
				}
			}
		}
	}

	return valueDisco, indexDisc, valuePart, indexPart
}

//deleteParticion borra un *NodoParticion del listado del disco con el index i
func deleteParticion(listadoParticion *[]*NodoParticion, i int) {
	// Remove the element at index i from a.
	copy((*listadoParticion)[i:], (*listadoParticion)[i+1:])           // Shift a[i+1:] left one index.
	(*listadoParticion)[len(*listadoParticion)-1] = nil                // Erase last element (write zero value).
	*listadoParticion = (*listadoParticion)[:len(*listadoParticion)-1] // Truncate slice.
}

//deleteDisco borra un *NodoDisco del listado de los montados
func deleteDisco(mountList *[]*NodoDisco, i int) {
	// Remove the element at index i from a.
	copy((*mountList)[i:], (*mountList)[i+1:])    // Shift a[i+1:] left one index.
	(*mountList)[len(*mountList)-1] = nil         // Erase last element (write zero value).
	*mountList = (*mountList)[:len(*mountList)-1] // Truncate slice.
}

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++MOUNT

func desplegarMount(mountList *[]*NodoDisco) {
	for _, val := range *mountList {
		for n, va := range *val.listadoParticion {
			fmt.Println("[Disco: " + val.path + ", Nombre Particion: " + va.nombre + ", Id: " + va.id + "]" + "[Posicion del arreglo del disco Part: " + strconv.Itoa(n) + "]")
		}
	}
}

//ComandoMount ejecuta el comando Mount
func ComandoMount(comando CONSTCOMANDO, mapa map[string]string, mountList *[]*NodoDisco) {
	fmt.Printf("address of slice abajo 2 %p  \n", &mountList)
	MBR := MBRStruct{}
	path := mapa["PATH"]
	nombre := mapa["NAME"]

	var nodoPart *NodoParticion
	nodoPart = new(NodoParticion)

	var nodoDis *NodoDisco
	nodoDis = new(NodoDisco)

	if ExtrarMBR(path, comando, &MBR) {
		//getEBRbyName
		var ebrFirst *EBRStruct
		ebrFirst = new(EBRStruct)
		ebrFirst = getFirstEBR(&MBR, path, comando)

		var EBR *EBRStruct //EBR
		EBR = new(EBRStruct)
		var ebrBool = false

		if ebrFirst != nil {
			ebrBool, EBR = getEBRbyName(path, ebrFirst, nombre, comando)
		} else {
			PrintError(ERROR, "Problemas al extraer el primer EBR del disco o quizas no exista una extendida")
			//return  //TODO:CAMBIO DE ULTIMO MOMENTO
		}

		//getParticionByName
		var partition *PartitionStruct //PartitionStruct
		partition = new(PartitionStruct)
		partitionBool, partition := getParticionByName(&MBR, nombre) //TODO: VER SI PUEDO MODIFICAR EL PARTITION ORIGINAL

		if !partitionBool && !ebrBool {
			PrintError(ERROR, "No existe una particion Primaria ni Logica con el nombre en el disco: [Nombre:"+nombre+", Primaria: "+strconv.FormatBool(partitionBool)+", Logica: "+strconv.FormatBool(ebrBool)+", Disco: "+path+"]")
			return
		}
		if partitionBool && ebrBool {
			PrintError(ERROR, "Existe dos particiones con el mismo nombre [Primaria y Logica]. No se puede montar la particion")
			return
		}

		//TODO: CONSTRUIR EL ID
		if partitionBool && partition != nil { //
			//tengo el partition
			nodoPart.partition = partition
			nodoPart.ebr = nil
			//nodoPart.id =
			if construirIDyMontar(mountList, path, nombre, nodoDis, nodoPart, comando) { //TODO: VER SI CUANDO HAY UN ERROR NO DEJO MODIFICA ALGO Y TENER QUE REVERTIRLO
				PrintAviso(comando, "Se termino el proceso de montar satisfactoriamente")
				return
			}
			PrintError(ERROR, "Existio un erro al montar la particion  [IdMontar: false]")
			return

		} else if ebrBool && EBR != nil {
			//tengo el EBR
			nodoPart.partition = nil
			nodoPart.ebr = EBR
			//nodoPart.id =
			if construirIDyMontar(mountList, path, nombre, nodoDis, nodoPart, comando) {
				PrintAviso(comando, "Se termino el proceso de montar satisfactoriamente")
				return
			}
			PrintError(ERROR, "Existio un erro al montar la particion  [IdMontar: false]")
			return

		} else {
			//==nil
			PrintError(ERROR, "Por algun error, No se puede montar la particion")
			return
		}

	} else {
		PrintError(ERROR, "Problemas al extraer el MBR del disco")
		return
	}
}

//isDiscoMontado me dice si el disco esta montado (si hay bool, letra string)
func isDiscoMontado(mountList *[]*NodoDisco, path string) (bool, string) {
	//tengo que recorrer los valores del map y buscar el path
	for _, value := range *mountList {
		if value.path == path {
			return true, value.letraDisco
		}
	}
	return false, ""
}

//getDiscoMontado obtiene un disco que ya esta montado con todo y sus atributos y su lista
func getDiscoMontado(mountList *[]*NodoDisco, path string) *NodoDisco {
	//tengo que recorrer los valores del map y buscar el path
	for _, value := range *mountList {
		if value.path == path {
			return value
		}
	}
	return nil
}

//isNamePartMontado me dice si una particion ya esta montada [true:si ya esta, false: no esta]
func isNamePartMontado(mountList *[]*NodoDisco, path string, namePart string) bool {
	fmt.Printf("address of slice %p  \n", &mountList)
	for _, value := range *mountList {
		if value.path == path {
			for _, val := range *value.listadoParticion {
				if val.nombre == namePart {
					return true
				}
			}
		}
	}
	return false
}

//getContadorPartDisco obtiene el contador total de particiones montadas del disco
func getContadorPartDisco(mountList *[]*NodoDisco, path string) int {
	for _, value := range *mountList {
		if value.path == path {
			return len(*value.listadoParticion)
		}
	}
	return 0
}

//getContadorDisco obtiene el contador total de discos en memoria
func getContadorDisco(mountList *[]*NodoDisco, path string) int {
	//tengo que recorrer los valores del map y buscar el path
	fmt.Printf("address of slice %p  \n", &mountList)
	return len(*mountList)
}

//construirIDyMontar contruye un id para Mount y monta las particiones en memoria
func construirIDyMontar(mountList *[]*NodoDisco, path string, namePart string, NewNodoDisco *NodoDisco, nodoPart *NodoParticion, comando CONSTCOMANDO) bool {

	fmt.Printf("address of slice abajo 3 %p  \n", &mountList)

	id := "vd"
	isMontadoDisk, letra := isDiscoMontado(mountList, path)
	if isMontadoDisk { //si esta montado ya tiene letra

		if !isNamePartMontado(mountList, path, namePart) { //sino esta montado lo monto

			var discoMontado *NodoDisco
			discoMontado = getDiscoMontado(mountList, path)

			if discoMontado != nil {

				PrintAviso(comando, "Montando una nueva particion a un disco ya existente en memoria")
				//contadorPart := getContadorPartDisco(mountList, path) //la ultima particion >= 1
				//contadorPart = contadorPart + 1              //la siguiente particion
				//id = id + letra + strconv.Itoa(contadorPart) //id+letra del disco montado + (+1)particion
				discoMontado.contador = discoMontado.contador + 1
				id = id + letra + strconv.Itoa(discoMontado.contador)
				nodoPart.id = id
				nodoPart.nombre = namePart
				nodoPart.letraDisco = letra
				var comodin = [10]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
				nodoPart.usuario = comodin
				nodoPart.contrasena = comodin
				//nodoDisco.letraDisco = letra
				*discoMontado.listadoParticion = append(*discoMontado.listadoParticion, nodoPart) //agrego la particion a la lista de particiones
				PrintAviso(comando, "Se monto exitosamente una nueva particion con el id y en el disco [NombrePart: "+namePart+", Id: "+id+", Disco: "+path+"]")

				fmt.Printf("address of slice abajo 4 %p  \n", &listaMontados)
				return true

			}
			PrintError(ERROR, "Error al extraer la informacion del disco montado para montar otra de sus particiones")
			return false

		} //si ya esta montado no se puede volver a montar
		PrintError(ERROR, "La particion [Nombre: "+namePart+"] del disco [Disco: "+path+"] ya esta montada en el")
		return false

	}
	//sino
	//CREAR NODODISCO //CON SU NODOPARTICION //VINCULARLOS
	PrintAviso(comando, "Creando nueva informacion de un disco para montar una de sus particiones")

	NewNodoDisco.path = path
	NewNodoDisco.listadoParticion = new([]*NodoParticion) //*[]*NodoParticion
	*NewNodoDisco.listadoParticion = make([]*NodoParticion, 0)

	//contadorDisk := getContadorDisco(mountList, path) //numero del disco
	letraNueva := letras[contadorDiscos] //extrae la letra (numero-letra anterior(0))
	contadorDiscos++
	id = id + letraNueva + strconv.Itoa(1) //TODO: revisar este contador
	nodoPart.id = id
	nodoPart.nombre = namePart
	nodoPart.letraDisco = letraNueva
	var comodin = [10]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	nodoPart.usuario = comodin
	nodoPart.contrasena = comodin
	NewNodoDisco.letraDisco = letraNueva
	NewNodoDisco.contador = 1
	*NewNodoDisco.listadoParticion = append(*NewNodoDisco.listadoParticion, nodoPart) //agrego la particion a la lista de particiones
	*mountList = append(*mountList, NewNodoDisco)                                     //agrego el disco
	PrintAviso(comando, "Se monto exitosamente una nueva particion con el id del disco [NombrePart: "+namePart+", Id: "+id+", Disco: "+path+"]")

	fmt.Printf("address of slice abajo 5 %p  \n", &listaMontados)
	return true

}

/*
//ordenarMap ordena el mapa por las keys [id]
func ordenarMap(mountMap map[string]NodoMontado) map[string]NodoMontado {

	nueva := make(map[string]NodoMontado)

	keys := make([]string, 0, len(mountMap)) //los id's ordenados
	for k := range mountMap {                //obtengo las llaves
		keys = append(keys, k) //y se los agrego al nuevo arreglo
	}
	sort.Strings(keys) //ordeno los strings (llaves)

	for _, k := range keys { //del arreglo ordena, obtengo el i,val
		fmt.Println(k, mountMap[k]) //imprimo el val y voy a buscar el val en el map
		nueva[k] = mountMap[k]
	}
	return nueva
}
*/

//getParticionByName obtiene una particion [principal] si existe con el mismo nombre
func getParticionByName(mbr *MBRStruct, name string) (bool, *PartitionStruct) {

	nombre := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	copy(nombre[:], name)

	for i := range mbr.Partition {
		if mbr.Partition[i].PartStatus != '0' {
			if mbr.Partition[i].PartType == 'P' {
				if mbr.Partition[i].PartName == nombre { //int = 48//byte = '0'  || int = 0// int = 0
					return true, &mbr.Partition[i] //TODO: VERIFICAR SI MODIFICO AL ORIGINAL
				}
			}
		}
	}
	return false, nil
}

//TODO: VERIFICAR QUE HACERLE CASTEO A LOS MBR O DATOS DEL DISCO NO ME PERJUDIQUEN
//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++LOGICAS

//ComandoFdisk ejecuta Fdisk
func ComandoFdisk(comando CONSTCOMANDO, mapa map[string]string, sizePartition int) {

	//SINO PROCEDO NORMAL [COMO PARA CREAR UNA PARTICION]
	//BUSCAR EL ARCHIVO//EXTRAER EL MBR//VER QUE VENGAN TODOS LOS PARAMETROS PUES [llenito]
	//VER SI TIENE DISPONIBLE UNA PARTICION LIBRE PARA CREAR OTRA [YA HAY 4 PARTITIONS OCUPADAS][si es logica no me interesa esto]
	//VER SI HAY ALGUNA DISPONIBLE [0 DESHABILITADA][4 PRINCIPALES]
	//VER QUE EL NOMBRE NO SE REPETIDA CON OTRA PARTITION

	//SWITCH TIPO
	//QUE TIPO DE PARTICION ESTOY CREANDO
	//SI ES EXTENDIDA VER QUE NO EXISTA OTRA YA [SI ES EXTENDIDA]
	//SI ES LOGICA VER SI YA EXISTE UNA EXTENDIDA
	//SI ES LOGICA VER QUE SEA MENOR QUE LA [EXTENDIDA-EBR]
	//--------------------------------------------------------------------------------------------------------------------
	//VER SI HAY ALGUNA DISPONIBLE [-1 DESHABILITADA][LOGICAS]
	//VER SI TIENE DISPONIBLE ESPACIO [PRIMER AJUSTE]
	//VER QUE EL NOMBRE NO SE REPETIDA CON OTRA PARTITION
	//SUS DEMAS RESTRICCIONES PROPIAS DEL CASO
	//VER SI TIENE DISPONIBLE ESPACIO [PRIMER AJUSTE][NO TIENE QUE SER MAYOR AL DISCO]

	//LLENAR LA INFORMACION DE UNA PARTICION EN PARTITION LUEGO VOLVER A GUARDAR EL MBR

	//TODO: VALIR REPETICION DE NOMBRES ENTRE TODOS Y NO SOLO LOS PRINCIPALES

	tamanio := TamanioTotal(comando, sizePartition, mapa["UNIT"])
	path := mapa["PATH"]
	tipo := mapa["TYPE"]
	name := mapa["NAME"]

	if tamanio != -1 {

		//ir a buscar el disco //ruta ya existe//leer el archivo//extraer mbr//validaciones......//todo ok?//llenar los datos de Partition en el MBR
		//...si es necesario ir a escribir al archivo EBR [logicas]

		MBR := MBRStruct{}
		if ExtrarMBR(path, comando, &MBR) {
			//fmt.Println("Asi esta el MBR extraido antes de trabajarlo")
			//fmt.Println(MBR) //--------------------------------

			if tipo == "P" || tipo == "E" {

				contadorLibres := ParticionesLibres(&MBR) //contador := MBR.ParticionesLibres()
				if contadorLibres > 0 {
					nombre := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
					copy(nombre[:], name)
					isRepetida := ParticionRepiteNombre(&MBR, nombre)
					if !isRepetida {

						totalDisco := int(MBR.MbrTamanio) - binary.Size(MBR)

						if tamanio <= totalDisco {

							if tipo == "P" {

								resultadoGuardarPrimaria := GuardarPrimaria(&MBR, tamanio, mapa, comando)

								if resultadoGuardarPrimaria {
									PrintAviso(comando, "La particion primaria se creo exitosamente")
									//fmt.Println("Asi quedo el MBR en memoria")
									//fmt.Println(MBR) //-------------------------------
									//fmt.Println("Extrayendo nuevamente el MBR del disco.....")
									//fmt.Println("Asi quedo el MBR en el disco")
									//mbr2 := MBRStruct{}
									//ExtrarMBR(mapa["PATH"], comando, &mbr2)
									//fmt.Println(mbr2)
									return
								}
								PrintError(ERROR, "La particion primaria no se pudo crear")
								return

							} else if tipo == "E" {

								//t == "E"
								if !ExisteExtendida(&MBR) {

									ebr := EBRStruct{}
									if tamanio > binary.Size(ebr) {

										//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
										//GuardarExtendida() //TODO: que escriba su ebr

										resultadoGuardarExtendida := GuardarExtendida(&MBR, tamanio, mapa, comando, &ebr)

										if resultadoGuardarExtendida {
											PrintAviso(comando, "La particion extendida se creo exitosamente")
											//fmt.Println("Asi quedo el MBR en memoria")
											//fmt.Println(MBR) //-------------------------------
											//fmt.Println("Extrayendo nuevamente el MBR del disco.....")
											//fmt.Println("Asi quedo el MBR en el disco")
											//mbr2 := MBRStruct{}
											//ExtrarMBR(mapa["PATH"], comando, &mbr2)
											//fmt.Println(mbr2)

											return
										}
										PrintError(ERROR, "La particion extendida no se pudo crear")
										return

										//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

									}
									PrintError(ERROR, "El tamanio de la extendida debe de ser mayor a su EBR inicial")
									return

								}
								PrintError(ERROR, "Ya existe una particion extendida en el disco, no se puede crear otra")
								return

							} else {
								PrintError(ERROR, "El tipo de particion no es correcta")
								return
							}

						} else {
							PrintError(ERROR, "El tamanio de la particion es mayor que el tamanio del Disco - MBR [Disco - MBR: "+strconv.Itoa(totalDisco)+", Particion:"+strconv.Itoa(tamanio)+"]")
							return
						}

					} else {
						PrintError(ERROR, "Ya existe una particion con este nombre ["+name+"]")
						return
					}

				} else {
					PrintError(ERROR, "Ya no hay espacio en el MBR para crear otra particion [P|E]")
					return
				}

			} else if tipo == "L" {

				if ExisteExtendida(&MBR) {
					ebr := EBRStruct{}

					if tamanio > binary.Size(ebr) {

						totalExt := int(TamanioExtendida(&MBR)) - binary.Size(ebr)
						if tamanio <= totalExt {
							var ebrComodin *EBRStruct
							ebrComodin = new(EBRStruct)
							ebrComodin = getFirstEBR(&MBR, mapa["PATH"], comando)

							if ebrComodin != nil {

								existName, resultEBR := getEBRbyName(mapa["PATH"], ebrComodin, name, comando) //[true,**][false,nil]
								if !existName {                                                               //TODO: VALIDAR NOMBRE

									resultPrimerAjusteLogica := int(getLogicalPrimerAjuste(&MBR, mapa["PATH"], comando, int64(tamanio)))
									if resultPrimerAjusteLogica != -1 { //TODO: LIBRE
										ejecucionLogica := EjecutarLogica(path, ebrComodin, int64(resultPrimerAjusteLogica), comando, int64(tamanio), mapa)
										//TODO: GRABAR
										//TODO: UNIR [ANTERIOR-SIGUIENTE]
										if ejecucionLogica {
											PrintAviso(comando, "La particion logica se creo exitosamente")

											//fmt.Println("PRIMER EBR")
											//fmt.Println(ebrComodin)
											//if getNextEBR(mapa["PATH"], ebrComodin, comando) == nil {
											//fmt.Println("***nil")
											//}

											return
										}
										PrintError(ERROR, "No se pudo crear la particion logica por algun error")
										return

									}
									PrintError(ERROR, "No hay suficiente espacio en la extendida para colocar la particion logica con el primer ajuste")
									return

								}
								PrintError(ERROR, "Ya existe una particion logica con este nombre [Nombre: "+name+", Nombre Particion Logica: "+string(resultEBR.PartName[:])+"]")
								return

							}
							PrintError(ERROR, "Error al extraer el primer EBR para trabajar sobre las logicas")
							return

						}
						PrintError(ERROR, "El tamanio de la particion logica es mayor que el tamanio de la particion extendida[Extendida - EBR: "+strconv.Itoa(totalExt)+", Particion Logica:"+strconv.Itoa(tamanio)+"]")
						return

					}
					PrintError(ERROR, "El tamanio de la logica debe de ser mayor a su EBR")
					return

				}
				PrintError(ERROR, "No existe una particion extendida en el disco, no se puede crear una particion logica")
				return

			} else {
				PrintError(ERROR, "El tipo no es correcto para proceder a crear la particion")
				return
			}

			//fmt.Println(MBR) //-------------------------------
		} else {
			PrintError(ERROR, "Problemas para extraer el MBR")
			return
		}

	} else {
		PrintError(comando, "Error al calcular el size en bytes")
		return
	}

}

//EjecutarLogica ejecuta las logicas [guardar-unir]
func EjecutarLogica(path string, firstEBR *EBRStruct, resultPrimerAjusteLogica int64, comando CONSTCOMANDO, sizePart int64, mapa map[string]string) bool {

	var prevEBR *EBRStruct
	prevEBR = new(EBRStruct)

	var ebrNext *EBRStruct   //== nil
	ebrNext = new(EBRStruct) //== &Student{"", 0}

	result := false

	prevEBR = getPreviousEBR(path, firstEBR, resultPrimerAjusteLogica, comando) //TODO: VER QUIEN ES EL firstEBR
	ebrNext = getNextEBR(path, prevEBR, comando)

	nextEBRStart := int64(-1)

	if firstEBR.PartStart == resultPrimerAjusteLogica {
		prevEBR = firstEBR
	}

	if prevEBR != nil {
		nextEBRStart = prevEBR.PartNext
		prevEBR.PartNext = resultPrimerAjusteLogica
		result = GuardarEBR(comando, prevEBR, path, int(prevEBR.PartStart))
	}

	if ebrNext != nil {
		nextEBRStart = ebrNext.PartStart
	}

	var nuevoEBR *EBRStruct
	nuevoEBR = new(EBRStruct)

	RellenarEBR(nuevoEBR, '1', resultPrimerAjusteLogica, sizePart, nextEBRStart, mapa)
	result = GuardarEBR(comando, nuevoEBR, path, int(nuevoEBR.PartStart))

	return result
}

//getLogicalPrimerAjuste obtiene el byte del primer ajuste para una logica [dentro de la extendida]
func getLogicalPrimerAjuste(mbr *MBRStruct, path string, comando CONSTCOMANDO, tamanioNuevaLogica int64) int64 {

	extendidaStart, extendidaEnd := InicioyFinExtendida(mbr)
	currentStart := extendidaStart
	partStart := int64(-1)
	partSize := int64(0)
	ebr := EBRStruct{}

	var previousEBR *EBRStruct
	previousEBR = new(EBRStruct)
	previousEBR = getFirstEBR(mbr, path, comando) //puede ser nil

	if previousEBR == nil {
		return -1
	}

	if previousEBR.PartSize > int64(binary.Size(ebr)) && previousEBR.PartNext >= -1 {
		currentStart = currentStart + previousEBR.PartSize
	}

	var nextEBR *EBRStruct //==nil
	nextEBR = new(EBRStruct)

	for {
		if nextEBR = getNextEBR(path, previousEBR, comando); nextEBR != nil {

			partStart = nextEBR.PartStart
			partSize = nextEBR.PartSize

			if (partStart - currentStart) >= tamanioNuevaLogica {
				break
			}

			currentStart = partStart + partSize
			previousEBR = nextEBR

		} else {
			break
		}
	}

	if currentStart <= (extendidaEnd - tamanioNuevaLogica) {
		return currentStart
	}
	return -1

}

//getEBRbyName obtiene un (bool, EBR) por su nombre
func getEBRbyName(path string, firstEBR *EBRStruct, name string, comando CONSTCOMANDO) (bool, *EBRStruct) {

	nombre := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	copy(nombre[:], name)

	var prevEBR *EBRStruct
	prevEBR = new(EBRStruct)
	prevEBR = firstEBR //==*EBRStruct

	var nextEBR *EBRStruct //==nil
	nextEBR = new(EBRStruct)

	if firstEBR.PartName == nombre {
		return true, firstEBR
	}

	for {
		if nextEBR = getNextEBR(path, prevEBR, comando); nextEBR != nil {

			if nextEBR.PartName == nombre {
				return true, nextEBR
			}
			prevEBR = nextEBR

		} else {
			break
		}
	}
	return false, nil
}

//getPreviousEBR obtiene el EBR anterior al que busco [buscado]
func getPreviousEBR(path string, firstEBR *EBRStruct, byteStart int64, comando CONSTCOMANDO) *EBRStruct {
	if byteStart == firstEBR.PartStart {
		return nil
	}

	var prevEBR *EBRStruct
	prevEBR = firstEBR

	var nextEBR *EBRStruct
	nextEBR = new(EBRStruct)

	for {
		if nextEBR = getNextEBR(path, prevEBR, comando); nextEBR != nil && nextEBR.PartStart < byteStart {
			prevEBR = nextEBR //TODO: verificar la condicion porque pareciera que voy para adelante pero luego no por una condicionalS
		} else {
			break
		}
	}
	return prevEBR
}

//getNextEBR obtiene el siguiente EBR del que busco [buscado]
func getNextEBR(path string, previousEBR *EBRStruct, comando CONSTCOMANDO) *EBRStruct {
	var ebrNext *EBRStruct   //== nil
	ebrNext = new(EBRStruct) //== &Student{"", 0}

	if previousEBR == nil || previousEBR.PartNext == -1 { //TODO: VERIFICAR QUE DEBO INICIALIZAR ANTES UN FINAL
		return nil
	}
	ExtrarEBR(path, comando, ebrNext, int(previousEBR.PartNext))
	return ebrNext
}

//getLastEBR no nil sino diferente al primero [firstEBR *EBRStruct]
func getLastEBR(path string, firstEBR *EBRStruct, comando CONSTCOMANDO) *EBRStruct {

	var prevEBR *EBRStruct
	prevEBR = firstEBR

	var nextEBR *EBRStruct
	nextEBR = new(EBRStruct)

	for {
		if nextEBR = getNextEBR(path, prevEBR, comando); nextEBR != nil {
			prevEBR = nextEBR //TODO: verificar la condicion porque pareciera que voy para adelante pero luego no por una condicionalS
		} else {
			break
		}
	}
	return prevEBR
}

//getFirstEBR obtiene el primer EBR
func getFirstEBR(mbr *MBRStruct, path string, comando CONSTCOMANDO) *EBRStruct {

	var primerEBR *EBRStruct   //==nil
	primerEBR = new(EBRStruct) //== &Student{"", 0}

	inicio, fin := InicioyFinExtendida(mbr)
	if inicio != -1 && fin != -1 {
		if ExtrarEBR(path, comando, primerEBR, int(inicio)) {
			return primerEBR
		}
		PrintError(ERROR, "Error al extraer el primer EBR")
		return nil
	}
	PrintError(ERROR, "No se pudo extraer el primer EBR por problemas al calcular el inicio y fin de la extendida")
	return nil

}

//RellenarEBR (&EBR) rellena los datos de un EBR [A ESTO SUMARLE EL binary.Size(EBR)]
func RellenarEBR(ebr *EBRStruct, statusPart byte, posStart int64, sizePart int64, posNext int64, mapa map[string]string) {

	nombreComodin := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}

	fit := [1]byte{'0'}
	copy(fit[:], mapa["FIT"])

	ebr.PartStatus = statusPart //en este caso indica que no tiene aun una particion logica [PORQUE SOLO ESTOY CREANDO EL EBR INICIAL]
	ebr.PartFit = fit[0]
	ebr.PartStart = posStart //donde inicia el EBR A ESTO SUMARLE EL binary.Size(EBR)
	ebr.PartSize = sizePart  //no tiene porque no es una logica es un EBR inicial
	ebr.PartNext = posNext
	ebr.PartName = nombreComodin
	copy(ebr.PartName[:], mapa["NAME"]) //no lleva nombre porque no hay particion en la EBR solo es la inicial
}

//InicioyFinExtendida byte donde inicia y byte donde termina la extendida [-1, -1] si hay error
func InicioyFinExtendida(mbr *MBRStruct) (int64, int64) {
	for i := range mbr.Partition {
		if mbr.Partition[i].PartStatus != '0' {
			if mbr.Partition[i].PartType == 'E' { //int = 48//byte = '0'  || int = 0// int = 0
				return mbr.Partition[i].PartStart, mbr.Partition[i].PartStart + mbr.Partition[i].PartSize
			}
		}
	}
	PrintError(ERROR, "No se pudo extraer el inicio y el fin de una extendida")
	return -1, -1
}

//Prue ddd
func Prue() {

	var cero int8
	s := &cero
	fmt.Println(binary.Size(s))    // 1
	fmt.Println(binary.Size(cero)) // 1
	fmt.Printf("% x\n", s)         //buffer.Bytes() //"0"  // c0000b600f
	fmt.Printf("% x\n", cero)      //buffer.Bytes() //"0"  // 0
	//fmt.Println(unsafe.Sizeof(s)) //int8
	//fmt.Println(unsafe.Sizeof(cero)) //= 0

	nulos2 := make([]byte, 25)
	//for i := range nulos2 {
	//	nulos2[i] = 9
	//}
	fmt.Println(binary.Size(nulos2))
	fmt.Printf("% x\n", nulos2) //buffer.Bytes() //"abcde"  //[00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00]

	var a [25]byte
	fmt.Println("++++")
	fmt.Println(binary.Size(a))
	fmt.Printf("% x\n", a) //0

	fmt.Println("++++")
	buf := new(bytes.Buffer)
	pi := make([]byte, 10)
	pi2 := &pi

	fmt.Printf("% x\n", pi2)

	fmt.Printf("% x\n", &pi)

	err := binary.Write(buf, binary.LittleEndian, pi2)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	//fmt.Printf("% x", buf.Bytes())

	fmt.Println(",,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,")
	tamanio := 300000
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	mbr := &MBRStruct{ //TODO: OBSERVAR EL &
		MbrTamanio:       int64(tamanio),
		MbrDiskSignature: int64(r.Int()),
		//Partition:
	}
	copy(mbr.MbrFechaCreacion[:], time.Now().Format("01-02-2006 15:04:05"))

	if mbr.Partition[0].PartStatus == 0 { //cero entero
		fmt.Println("1 es igual a 0")
	}
	if mbr.Partition[0].PartStatus == 48 { //cero en byte
		fmt.Println("1 es igual a 48")
	}

	if mbr.Partition[0].PartStatus == '0' { //valor '0'
		fmt.Println("1 VALOR es igual a 0")
	}

	for i := 0; i < 4; i++ {
		mbr.Partition[i].PartStatus = '0'
		mbr.Partition[i].PartType = '0'
		mbr.Partition[i].PartFit = '0'
		mbr.Partition[i].PartStart = -1
		mbr.Partition[i].PartSize = 0
		for n := range mbr.Partition[i].PartName {
			mbr.Partition[i].PartName[n] = '0'
		}
	}

	if mbr.Partition[0].PartStatus == 0 {
		fmt.Println("2 es igual a 0")
	}
	if mbr.Partition[0].PartStatus == 48 {
		fmt.Println("2 es igual a 48")
	}

	if mbr.Partition[0].PartStatus == '0' {
		fmt.Println("2 VALOR es igual a 0")
	}

	//MISMO TAMANIO PARA LAS 3 OPCIONES DE STRUCTS
	//LLENO
	fmt.Println(mbr)
	fmt.Println(binary.Size(mbr))
	/*
		//PUNTERO
		var mbr2 *MBRStruct   // mbr2 == nil
		mbr2 = new(MBRStruct) // mbr2 == &Student{"", 0}
		fmt.Println(binary.Size(mbr2))
		fmt.Println("puntero:")
		mbr2.MbrTamanio = 999999
		fmt.Println(mbr2)

		//VARIABLE SIN INICIALIZAR
		var mbr3 MBRStruct
		fmt.Println(binary.Size(mbr3))
		fmt.Println("variable SIN inicializar:")
		mbr3.MbrTamanio = 99999
		fmt.Println(mbr3)

		//VARIABLE INICIADA
		mbr4 := MBRStruct{}
		mbr4.MbrTamanio = 99999
		fmt.Println("variable inicializada:")
		fmt.Println(mbr4)
	*/
	//PUNTERO
	fmt.Println("============")
	var ebr *EBRStruct   // mbr2 == nil
	ebr = new(EBRStruct) // mbr2 == &Student{"", 0} // &mbr2
	fmt.Println(binary.Size(ebr))
	fmt.Println("puntero inicializado:")
	ebr.PartSize = 999999
	fmt.Println(ebr) //si lo quiero modificar solo lo paso asi

	fmt.Println("direccion antes: puntero original") //0xc00000e038 ES LA MISMA
	fmt.Println(&ebr)
	//ebr = RellenarEBR('0', '0', 0, 0, -1, "abcd1") //==(&ebr3) direccion//crear una nueva varibale puntero e igualarlo
	//ebr = nil
	//RellenarEBR(ebr, '0', '0', 0, 0, -1, "abcd1") //==(&ebr3) direccion//crear una nueva varibale puntero e igualarlo
	fmt.Println("variable rellenada:")
	fmt.Println(ebr)                                               //==(&ebr3) direccion
	fmt.Println("direccion despues de rellenar: puntero original") //0xc00000e038 ES LA MISMA
	fmt.Println(&ebr)
	ebr.PartNext = 1
	fmt.Println("variable rellenada y  luego modificada:")
	fmt.Println(ebr) //==(&ebr3) direccion

	//VARIABLE SIN INICIALIZAR, sus valores empiezan en 0
	var ebr2 EBRStruct
	fmt.Println(binary.Size(ebr2))
	fmt.Println("variable SIN inicializar:")
	ebr2.PartSize = 99999
	fmt.Println(ebr2)

	//VARIABLE INICIADA
	ebr3 := EBRStruct{}
	ebr3.PartSize = 99999
	fmt.Println("variable inicializada:")
	fmt.Println(ebr3)
	//no se puede porque es un cajon(puntero) igualarselo a una variable
	//&ebr3 = RellenarEBR('0', '0', 0, 0, -1, "abcd1") //==(&ebr3) direccion//crear una nueva varibale puntero e igualarlo
	/*
		var nodoDis *NodoDisco
		nodoDis = new(NodoDisco)

		nodoDis.path = "asdfasdfasd"
		nodoDis.letraDisco = "a"
		nodoDis.listadoParticion = make([]NodoParticion, 0)
		nodoPar := NodoParticion{nombre: "mi nombre"}
		nodoDis.listadoParticion = append(nodoDis.listadoParticion, nodoPar)
		nodoDis.listadoParticion = append(nodoDis.listadoParticion, nodoPar)
		nodoDis.listadoParticion = append(nodoDis.listadoParticion, nodoPar)
		nodoDis.listadoParticion = append(nodoDis.listadoParticion, nodoPar)
		nodoDis.listadoParticion = append(nodoDis.listadoParticion, nodoPar)
		nodoDis.listadoParticion = append(nodoDis.listadoParticion, nodoPar)
		nodoDis.listadoParticion = append(nodoDis.listadoParticion, nodoPar)
		nodoDis.listadoParticion = append(nodoDis.listadoParticion, nodoPar)
		nodoDis.listadoParticion = append(nodoDis.listadoParticion, nodoPar)

		fmt.Println("Nodo disco")
		fmt.Println(nodoDis)
	*/
	fmt.Println("LETRASS")
	fmt.Println(letras[0])
	fmt.Println(letras[1])
	fmt.Println(letras[2])
	fmt.Println(letras[3])
	fmt.Println(letras)
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

//GuardarPrimaria guarda una particion primaria
func GuardarPrimaria(mbr *MBRStruct, tamanio int, mapa map[string]string, comando CONSTCOMANDO) bool {

	startComodin := GetPrimerAjusteInicio(mbr, tamanio)

	if int(startComodin) == -1 {
		PrintError(ERROR, "No se encontro un espacio libre del tamaño de la particion dentro del disco")
		return false
	}

	status := [1]byte{'1'}

	tipo := [1]byte{'0'}
	copy(tipo[:], mapa["TYPE"])

	fit := [1]byte{'0'}
	copy(fit[:], mapa["FIT"])

	nombre := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	copy(nombre[:], mapa["NAME"])

	particion := PartitionStruct{PartStatus: status[0], PartType: tipo[0], PartFit: fit[0], PartStart: startComodin, PartSize: int64(tamanio), PartName: nombre}

	resultadoAgregar := AgregarParticionAlMBR(mbr, &particion)

	if resultadoAgregar == -1 {
		PrintError(ERROR, "No se puede agregar la particion al MBR del disco, por algun inconveniente")
		return false
	}

	resultadoGuardarMBR := GuardarMBR(comando, mbr, mapa["PATH"])

	if resultadoGuardarMBR {
		PrintAviso(comando, "Particion creada correctamente [Nombre: "+mapa["NAME"]+", Disco: "+mapa["PATH"]+"]")
		return true
	}
	PrintError(ERROR, "Error al crear la particion [Nombre: "+mapa["NAME"]+", Disco: "+mapa["PATH"]+"]")
	return false

}

//GuardarExtendida guarda una particion extendia y escribe el ebr inicial
func GuardarExtendida(mbr *MBRStruct, tamanio int, mapa map[string]string, comando CONSTCOMANDO, ebr *EBRStruct) bool {

	startComodin := GetPrimerAjusteInicio(mbr, tamanio)

	if int(startComodin) == -1 {
		PrintError(ERROR, "No se encontro un espacio libre del tamaño de la particion dentro del disco")
		return false
	}

	//TODO: talvez poder mejorar como se lleno el EBR
	status := [1]byte{'1'}

	tipo := [1]byte{'0'}
	copy(tipo[:], mapa["TYPE"])

	fit := [1]byte{'0'}
	copy(fit[:], mapa["FIT"])

	nombre := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	copy(nombre[:], mapa["NAME"])

	particion := PartitionStruct{PartStatus: status[0], PartType: tipo[0], PartFit: fit[0], PartStart: startComodin, PartSize: int64(tamanio), PartName: nombre}

	//TODO: VERIFICAR ESTO
	//NO ES NECESARIO GUARDAR UNA COPIA PORQUE SI EXISTE ALGUN ERROR EL DISCO QUEDA COMPLETAMENTE IGUAL A COMO ESTABA
	LLenarEBRInicial(ebr, startComodin)

	resultadoAgregar := AgregarParticionAlMBR(mbr, &particion)

	if resultadoAgregar == -1 {
		PrintError(ERROR, "No se pudo agregar la particion al MBR del disco, por algun inconveniente")
		return false
	}

	resultadoGuardarMBR := GuardarMBR(comando, mbr, mapa["PATH"])

	if !resultadoGuardarMBR {
		PrintError(ERROR, "Error al crear la particion [Nombre: "+mapa["NAME"]+", Disco: "+mapa["PATH"]+", BoolMBR:"+strconv.FormatBool(resultadoGuardarMBR)+"]")
		return false
	}

	resultadoGuardarEBR := GuardarEBR(comando, ebr, mapa["PATH"], int(ebr.PartStart))

	if !resultadoGuardarEBR {
		PrintError(ERROR, "Error al crear la particion [Nombre: "+mapa["NAME"]+", Disco: "+mapa["PATH"]+", BoolEBR:"+strconv.FormatBool(resultadoGuardarEBR)+"]")
		return false
	}

	PrintAviso(comando, "Particion creada correctamente [Nombre: "+mapa["NAME"]+", Disco: "+mapa["PATH"]+"]")

	fmt.Println("Extrayendo el EBR inicial desde el archivo con startComodin")
	eb := EBRStruct{}
	ExtrarEBR(mapa["PATH"], comando, &eb, int(startComodin))

	fmt.Println("Extrayendo el EBR inicial desde el archivo con ebr.PartStart ebr que llenamos y que escribimo en el disco")
	eb2 := EBRStruct{}
	ExtrarEBR(mapa["PATH"], comando, &eb2, int(ebr.PartStart))

	return true

}

//TODO: br.PartStart EN DONDE INICIA LA EXT O DESPUES DEL EBR INICIAL [startComodin + binary.Size(EBR)]?
//TODO: SE DEFINIO QUE DONDE INICIA LA EXT

//LLenarEBRInicial alista un EBR inicial para guardarlo
func LLenarEBRInicial(ebr *EBRStruct, startComodin int64) {

	nombre := [16]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}

	ebr.PartStatus = '0' //en este caso indica que no tiene aun una particion logica [PORQUE SOLO ESTOY CREANDO EL EBR INICIAL]
	ebr.PartFit = '0'
	ebr.PartStart = startComodin //donde inicia la particion EXTENDIDA y el EBR
	ebr.PartSize = int64(0)      //no tiene porque no es una logica es un EBR inicial
	ebr.PartNext = int64(-1)
	ebr.PartName = nombre //no lleva nombre porque no hay particion en la EBR solo es la inicial
}

//GuardarMBR guarda el MBR en el disco
func GuardarMBR(comando CONSTCOMANDO, mbr *MBRStruct, path string) bool {
	// OpenFile with more options. Last param is the permission mode
	// Second param is the attributes when opening
	//--------------------------------------------------------
	// os.O_RDONLY // Read only
	// os.O_WRONLY // Write only
	// os.O_RDWR // Read and write
	// os.O_APPEND // Append to end of file
	// os.O_CREATE // Create is none exist
	// os.O_TRUNC // Truncate file when opening
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		PrintError(ERROR, "Error al abrir el archivo")
		log.Fatal(err)
		return false
	}

	bufferBinarioMBR := new(bytes.Buffer)
	if binary.Write(bufferBinarioMBR, binary.BigEndian, mbr); err != nil {
		PrintError(ERROR, "Error al hacer la conversion a binario")
		fmt.Println("binary.Write failed:", err)
		file.Close()
		return false
	}

	if WriteBytes2(file, bufferBinarioMBR.Bytes(), 0, 0, comando) {
		PrintAviso(comando, "Se escribio el MBR correctamente")
		return true
	}
	PrintError(ERROR, "Error al escribir el MBR en el disco")
	return false
}

//GuardarEBR guarda un EBR en el disco
func GuardarEBR(comando CONSTCOMANDO, ebr *EBRStruct, path string, posicion int) bool {
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		PrintError(ERROR, "Error al abrir el archivo")
		log.Fatal(err)
		return false
	}

	bufferBinarioEBR := new(bytes.Buffer)
	if binary.Write(bufferBinarioEBR, binary.BigEndian, ebr); err != nil {
		PrintError(ERROR, "Error al hacer la conversion a binario")
		fmt.Println("binary.Write failed:", err)
		file.Close()
		return false
	}

	if WriteBytes2(file, bufferBinarioEBR.Bytes(), posicion, 0, comando) {
		PrintAviso(comando, "Se escribio el EBR correctamente")
		return true
	}
	PrintError(ERROR, "Error al escribir el EBR en el disco")
	return false
}

//GuardarSB guarda un SB en el disco
func GuardarSB(comando CONSTCOMANDO, sb *SuperBootStruct, path string, posicion int, isCopia bool) bool {
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		PrintError(ERROR, "Error al abrir el archivo")
		log.Fatal(err)
		return false
	}

	bufferBinarioEBR := new(bytes.Buffer)
	if binary.Write(bufferBinarioEBR, binary.BigEndian, sb); err != nil {
		PrintError(ERROR, "Error al hacer la conversion a binario")
		fmt.Println("binary.Write failed:", err)
		file.Close()
		return false
	}

	if WriteBytes2(file, bufferBinarioEBR.Bytes(), posicion, 0, comando) {
		if isCopia {
			PrintAviso(comando, "Se escribio el SB copia correctamente")
		} else {
			PrintAviso(comando, "Se escribio el SB correctamente")
		}
		return true
	}
	if isCopia {
		PrintError(ERROR, "Error al escribir el SB copia en el disco")
	} else {
		PrintError(ERROR, "Error al escribir el SB en el disco")
	}
	return false
}

//GuardarBitmap guarda un bitmap en el disco
func GuardarBitmap(comando CONSTCOMANDO, bitmap []byte, path string, posicion int, nombreBitmap string) bool {
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		PrintError(ERROR, "Error al abrir el archivo")
		log.Fatal(err)
		return false
	}

	if WriteBytes2(file, bitmap, posicion, 0, comando) {
		PrintAviso(comando, "Se escribio el Bitmap correctamente de [Bitmap: "+nombreBitmap+"]")
		return true
	}
	PrintError(ERROR, "Error al escribir el Bitmap correctamente de [Bitmap: "+nombreBitmap+"]")
	return false
}

//GuardarAVD guarda un EBR en el disco
func GuardarAVD(comando CONSTCOMANDO, avd *AVDStruct, path string, posicion int) bool {
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		PrintError(ERROR, "Error al abrir el archivo")
		log.Fatal(err)
		return false
	}

	bufferBinarioAVD := new(bytes.Buffer)
	if binary.Write(bufferBinarioAVD, binary.BigEndian, avd); err != nil {
		PrintError(ERROR, "Error al hacer la conversion a binario")
		fmt.Println("binary.Write failed:", err)
		file.Close()
		return false
	}

	if WriteBytes2(file, bufferBinarioAVD.Bytes(), posicion, 0, comando) {
		PrintAviso(comando, "Se escribio el AVD correctamente")
		return true
	}
	PrintError(ERROR, "Error al escribir el AVD en el disco")
	return false
}

//GuardarDD guarda un EBR en el disco
func GuardarDD(comando CONSTCOMANDO, dd *DDStruct, path string, posicion int) bool {
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		PrintError(ERROR, "Error al abrir el archivo")
		log.Fatal(err)
		return false
	}

	bufferBinarioDD := new(bytes.Buffer)
	if binary.Write(bufferBinarioDD, binary.BigEndian, dd); err != nil {
		PrintError(ERROR, "Error al hacer la conversion a binario")
		fmt.Println("binary.Write failed:", err)
		file.Close()
		return false
	}

	if WriteBytes2(file, bufferBinarioDD.Bytes(), posicion, 0, comando) {
		PrintAviso(comando, "Se escribio el DD correctamente")
		return true
	}
	PrintError(ERROR, "Error al escribir el DD en el disco")
	return false
}

//GuardarInodo guarda un EBR en el disco
func GuardarInodo(comando CONSTCOMANDO, inodo *InodoStruct, path string, posicion int) bool {
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		PrintError(ERROR, "Error al abrir el archivo")
		log.Fatal(err)
		return false
	}

	bufferBinarioInodo := new(bytes.Buffer)
	if binary.Write(bufferBinarioInodo, binary.BigEndian, inodo); err != nil {
		PrintError(ERROR, "Error al hacer la conversion a binario")
		fmt.Println("binary.Write failed:", err)
		file.Close()
		return false
	}

	if WriteBytes2(file, bufferBinarioInodo.Bytes(), posicion, 0, comando) {
		PrintAviso(comando, "Se escribio el Inodo correctamente")
		return true
	}
	PrintError(ERROR, "Error al escribir el Inodo en el disco")
	return false
}

//GuardarBD guarda un EBR en el disco
func GuardarBD(comando CONSTCOMANDO, bd *BloqueDeDatosStruct, path string, posicion int) bool {
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		PrintError(ERROR, "Error al abrir el archivo")
		log.Fatal(err)
		return false
	}

	bufferBinarioBD := new(bytes.Buffer)
	if binary.Write(bufferBinarioBD, binary.BigEndian, bd); err != nil {
		PrintError(ERROR, "Error al hacer la conversion a binario")
		fmt.Println("binary.Write failed:", err)
		file.Close()
		return false
	}

	if WriteBytes2(file, bufferBinarioBD.Bytes(), posicion, 0, comando) {
		PrintAviso(comando, "Se escribio el BD correctamente")
		return true
	}
	PrintError(ERROR, "Error al escribir el BD en el disco")
	return false
}

//GuardarLog guarda un EBR en el disco
func GuardarLog(comando CONSTCOMANDO, bitacora *BitacoraStruct, path string, posicion int) bool {
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		PrintError(ERROR, "Error al abrir el archivo")
		log.Fatal(err)
		return false
	}

	bufferBinarioLog := new(bytes.Buffer)
	if binary.Write(bufferBinarioLog, binary.BigEndian, bitacora); err != nil {
		PrintError(ERROR, "Error al hacer la conversion a binario")
		fmt.Println("binary.Write failed:", err)
		file.Close()
		return false
	}

	if WriteBytes2(file, bufferBinarioLog.Bytes(), posicion, 0, comando) {
		PrintAviso(comando, "Se escribio la Bitacora correctamente")
		return true
	}
	PrintError(ERROR, "Error al escribir la Bitacora en el disco")
	return false
}

//AgregarParticionAlMBR agregar una particion al MBR
func AgregarParticionAlMBR(mbr *MBRStruct, particion *PartitionStruct) int {
	for i := 0; i < 4; i++ {
		if mbr.Partition[i].PartStart == -1 { //TODO: VERIFICAR QUE SEA UNA CONDICION CORRECTA [-1]
			mbr.Partition[i] = *particion //TODO: VERIFICAR QUE SI LE ESTOY PASANDO LOS DATOS CORRECTOS*
			OrdenarMBRParticiones(mbr)
			return 1
		}
	}
	return 0
}

//OrdenarMBRParticiones ordena las particiones luego de insertarlas
func OrdenarMBRParticiones(mbr *MBRStruct) {

	var tmp PartitionStruct

	for i := 1; i < 4; i++ {
		for j := 0; j < 3; j++ {
			if mbr.Partition[j].PartStart > mbr.Partition[j+1].PartStart {
				tmp = mbr.Partition[j]
				mbr.Partition[j] = mbr.Partition[j+1]
				mbr.Partition[j+1] = tmp
			}
			if mbr.Partition[j].PartStart == -1 {
				tmp = mbr.Partition[j]
				mbr.Partition[j] = mbr.Partition[j+1]
				mbr.Partition[j+1] = tmp
			}
		}
	}
}

//GetPrimerAjusteInicio obtiene el byte de inicio despues de aplicar el Primer Ajuste
func GetPrimerAjusteInicio(mbr *MBRStruct, tamanioParticionNueva int) int64 { //TODO: TOMAR EN CUENTA QUE AQUI YA ESTA EN ORDEN

	//currentStart := int64(binary.Size(mbr))
	//partStart := int64(-1)
	//partSize := int64(0)
	//
	//numPartitions := ParticionesOcupadas(mbr)
	//
	//for i := 0; i < numPartitions; i++ {
	//	partStart = mbr.Partition[i].PartStart
	//	partSize = mbr.Partition[i].PartSize
	//
	//	if (partStart - currentStart) >= int64(tamanioParticionNueva) {
	//		break
	//	}
	//	currentStart = partStart + partSize
	//}
	//
	//if currentStart <= (mbr.MbrTamanio - int64(tamanioParticionNueva)) {
	//	return currentStart
	//}
	//return -1

	start := int64(binary.Size(mbr)) //posArchivo
	ultimoEnd := int64(binary.Size(mbr))
	end := int64(0)

	listaLibres := make([]Libre, 0)

	for i := range mbr.Partition {
		if mbr.Partition[i].PartStatus == '1' {
			if mbr.Partition[i].PartStart != -1 { //.PartStatus == 49 (1)(si ocupado) || == 48 (0)(no ocupado)
				end = mbr.Partition[i].PartStart
				if (end - start) > 0 { //hay espacio libre
					libre := Libre{
						Lstart: int(start),
						Lend:   int(end),
					}
					listaLibres = append(listaLibres, libre) //texto += "Tamanio " + strconv.Itoa(int(libre.Lend-libre.Lstart)) + "<br/>"
					if tamanioParticionNueva <= (libre.Lend - libre.Lstart) {
						return int64(libre.Lstart)
					}
				}
				start = mbr.Partition[i].PartStart + mbr.Partition[i].PartSize

				ultimoEnd = mbr.Partition[i].PartStart + mbr.Partition[i].PartSize
			}
		}
	}

	if ultimoEnd < mbr.MbrTamanio {
		libre := Libre{
			Lstart: int(ultimoEnd),
			Lend:   int(mbr.MbrTamanio),
		}
		listaLibres = append(listaLibres, libre)
		if tamanioParticionNueva <= (libre.Lend - libre.Lstart) {
			return int64(libre.Lstart)
		}
	}
	return -1

}

//ParticionesOcupadas te indica cuantas particiones ocupadas hay
func ParticionesOcupadas(mbr *MBRStruct) int {
	var contador = 0
	for i := range mbr.Partition { //DEPENDE DEL TIPO DEL PARAMETRO
		if mbr.Partition[i].PartStatus != '0' { //int = 48//byte = '0'  || int = 0// int = 0
			//mbr.Partition[i].PartStatus = '9'
			contador++
		}
	}
	return contador
}

//ParticionesLibres te indica cuantas particiones libres hay
func ParticionesLibres(mbr *MBRStruct) int {
	var contador = 0
	for i := range mbr.Partition { //DEPENDE DEL TIPO DEL PARAMETRO
		if mbr.Partition[i].PartStatus == '0' { //int = 48//byte = '0'  || int = 0// int = 0
			//mbr.Partition[i].PartStatus = '9'
			contador++
		}
	}
	return contador
}

//ParticionRepiteNombre ya existe el nombre en las particiones activas
func ParticionRepiteNombre(mbr *MBRStruct, nombre [16]byte) bool {
	for i := range mbr.Partition {
		if mbr.Partition[i].PartStatus != '0' {
			if mbr.Partition[i].PartName == nombre { //int = 48//byte = '0'  || int = 0// int = 0
				return true
			}
		}
	}
	return false
}

//ExisteExtendida existe extendida
func ExisteExtendida(mbr *MBRStruct) bool {
	for i := range mbr.Partition {
		if mbr.Partition[i].PartStatus != '0' {
			if mbr.Partition[i].PartType == 'E' { //int = 48//byte = '0'  || int = 0// int = 0
				return true
			}
		}
	}
	return false
}

//TamanioExtendida tamanio de la extendida
func TamanioExtendida(mbr *MBRStruct) int64 {
	for i := range mbr.Partition {
		if mbr.Partition[i].PartStatus != '0' {
			if mbr.Partition[i].PartType == 'E' { //int = 48//byte = '0'  || int = 0// int = 0
				return mbr.Partition[i].PartSize
			}
		}
	}
	return 0
}

//EXCELENTE CLASE DE APUNTADORES+++++++++++++++++++
//te envio una copia [valor]
//NO el original la base

//ExtrarMBR extrae un MBR del disco
func ExtrarMBR(path string, comando CONSTCOMANDO, mbr *MBRStruct) bool { //paso una copia //paso el original la base

	//tengo una copia
	//tengo el original la base

	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
		return false
	}

	sizeBytesMBR := binary.Size(mbr) //TODO: VER QUE ESTE CASTEO NO AFECTE

	data := ReadBytes(file, sizeBytesMBR, 0, 0, comando)

	if data != nil {

		buffer := bytes.NewBuffer(data)

		//fmt.Println("Data:")
		//fmt.Println(data)

		err = binary.Read(buffer, binary.BigEndian, mbr) //* //& //NADA [la direcion de la otra var]][ir a buscarla]
		if err != nil {
			log.Fatal("binary.Read failed", err)
			file.Close()
			return false
		}

		//fmt.Println("MBR:")
		//fmt.Println(mbr)
		return true

	}
	PrintError(comando, "Error al extraer el MBR del archivo")
	return false

}

//ExtrarEBR extrae un EBR del disco
func ExtrarEBR(path string, comando CONSTCOMANDO, ebr *EBRStruct, posicion int) bool { //paso una copia //paso el original la base

	//tengo una copia
	//tengo el original la base

	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
		return false
	}

	sizeBytesEBR := binary.Size(ebr) //TODO: VER QUE ESTE CASTEO NO AFECTE

	data := ReadBytes(file, sizeBytesEBR, posicion, 0, comando)

	if data != nil {

		buffer := bytes.NewBuffer(data)

		//fmt.Println("Data:")
		//fmt.Println(data)

		err = binary.Read(buffer, binary.BigEndian, ebr) //* //& //NADA [la direcion de la otra var]][ir a buscarla]
		if err != nil {
			log.Fatal("binary.Read failed", err)
			file.Close()
			PrintError(ERROR, "Error al momento de decodificar el ebr")
			return false
		}

		//fmt.Println("EBR:")
		//fmt.Println(ebr)
		return true

	}
	PrintError(comando, "Error al extraer el EBR del archivo")
	return false

}

//ExtrarSB extrae un SB del disco
func ExtrarSB(path string, comando CONSTCOMANDO, sb *SuperBootStruct, posicion int) bool { //paso una copia //paso el original la base

	//tengo una copia
	//tengo el original la base
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
		return false
	}

	sizeBytesSB := binary.Size(sb) //TODO: VER QUE ESTE CASTEO NO AFECTE

	data := ReadBytes(file, sizeBytesSB, posicion, 0, comando)

	if data != nil {

		buffer := bytes.NewBuffer(data)

		//fmt.Println("Data:")
		//fmt.Println(data)

		err = binary.Read(buffer, binary.BigEndian, sb) //* //& //NADA [la direcion de la otra var]][ir a buscarla]
		if err != nil {
			log.Fatal("binary.Read failed", err)
			file.Close()
			PrintError(ERROR, "Error al momento de decodificar el SB")
			return false
		}

		//fmt.Println("SB:")
		//fmt.Println(sb)
		return true

	}
	PrintError(comando, "Error al extraer el SB del archivo")
	return false

}

//ExtrarBitmap extrae un bitmap completo
func ExtrarBitmap(path string, comando CONSTCOMANDO, start int, tamanio int, bitmap string) (bool, []byte) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
		return false, nil
	}

	data := ReadBytes(file, tamanio, start, 0, comando)
	if data != nil {
		PrintAviso(comando, "Se extrajo correctamente el Bitmap de ["+bitmap+"]")
		return true, data
	}
	PrintError(comando, "Error al extraer el Bitmap de ["+bitmap+"]")
	return false, nil

}

//ExtrarAVD extrae un AVD del disco
func ExtrarAVD(path string, comando CONSTCOMANDO, avd *AVDStruct, posicion int) bool { //paso una copia //paso el original la base

	//tengo una copia
	//tengo el original la base
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
		return false
	}

	sizeBytesAVD := binary.Size(avd) //TODO: VER QUE ESTE CASTEO NO AFECTE

	data := ReadBytes(file, sizeBytesAVD, posicion, 0, comando)

	if data != nil {

		buffer := bytes.NewBuffer(data)

		//fmt.Println("Data:")
		//fmt.Println(data)

		err = binary.Read(buffer, binary.BigEndian, avd) //* //& //NADA [la direcion de la otra var]][ir a buscarla]
		if err != nil {
			log.Fatal("binary.Read failed", err)
			file.Close()
			PrintError(ERROR, "Error al momento de decodificar el AVD")
			return false
		}

		//fmt.Println("AVD:")
		//fmt.Println(avd)
		return true

	}
	PrintError(comando, "Error al extraer el AVD del archivo")
	return false

}

//ExtrarDD extrae un DD del disco
func ExtrarDD(path string, comando CONSTCOMANDO, dd *DDStruct, posicion int) bool { //paso una copia //paso el original la base

	//tengo una copia
	//tengo el original la base
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
		return false
	}

	sizeBytesDD := binary.Size(dd) //TODO: VER QUE ESTE CASTEO NO AFECTE

	data := ReadBytes(file, sizeBytesDD, posicion, 0, comando)

	if data != nil {

		buffer := bytes.NewBuffer(data)

		//fmt.Println("Data:")
		//fmt.Println(data)

		err = binary.Read(buffer, binary.BigEndian, dd) //* //& //NADA [la direcion de la otra var]][ir a buscarla]
		if err != nil {
			log.Fatal("binary.Read failed", err)
			file.Close()
			PrintError(ERROR, "Error al momento de decodificar el DD")
			return false
		}

		//fmt.Println("DD:")
		//fmt.Println(dd)
		return true

	}
	PrintError(comando, "Error al extraer el DD del archivo")
	return false

}

//ExtrarInodo extrae un Inodo del disco
func ExtrarInodo(path string, comando CONSTCOMANDO, inodo *InodoStruct, posicion int) bool { //paso una copia //paso el original la base

	//tengo una copia
	//tengo el original la base
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
		return false
	}

	sizeBytesInodo := binary.Size(inodo) //TODO: VER QUE ESTE CASTEO NO AFECTE

	data := ReadBytes(file, sizeBytesInodo, posicion, 0, comando)

	if data != nil {

		buffer := bytes.NewBuffer(data)

		//fmt.Println("Data:")
		//fmt.Println(data)

		err = binary.Read(buffer, binary.BigEndian, inodo) //* //& //NADA [la direcion de la otra var]][ir a buscarla]
		if err != nil {
			log.Fatal("binary.Read failed", err)
			file.Close()
			PrintError(ERROR, "Error al momento de decodificar el Inodo")
			return false
		}

		//fmt.Println("Inodo:")
		//fmt.Println(inodo)
		return true

	}
	PrintError(comando, "Error al extraer el Inodo del archivo")
	return false

}

//ExtrarBD extrae un BD del disco
func ExtrarBD(path string, comando CONSTCOMANDO, bd *BloqueDeDatosStruct, posicion int) bool { //paso una copia //paso el original la base

	//tengo una copia
	//tengo el original la base
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
		return false
	}

	sizeBytesBD := binary.Size(bd) //TODO: VER QUE ESTE CASTEO NO AFECTE

	data := ReadBytes(file, sizeBytesBD, posicion, 0, comando)

	if data != nil {

		buffer := bytes.NewBuffer(data)

		//fmt.Println("Data:")
		//fmt.Println(data)

		err = binary.Read(buffer, binary.BigEndian, bd) //* //& //NADA [la direcion de la otra var]][ir a buscarla]
		if err != nil {
			log.Fatal("binary.Read failed", err)
			file.Close()
			PrintError(ERROR, "Error al momento de decodificar el BD")
			return false
		}

		//fmt.Println("BD:")
		//fmt.Println(bd)
		return true

	}
	PrintError(comando, "Error al extraer el BD del archivo")
	return false

}

//ExtrarLog extrae un Log del disco
func ExtrarLog(path string, comando CONSTCOMANDO, bitacora *BitacoraStruct, posicion int) bool { //paso una copia //paso el original la base

	//tengo una copia
	//tengo el original la base
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
		return false
	}

	sizeBytesBitacora := binary.Size(bitacora) //TODO: VER QUE ESTE CASTEO NO AFECTE

	data := ReadBytes(file, sizeBytesBitacora, posicion, 0, comando)

	if data != nil {

		buffer := bytes.NewBuffer(data)

		//fmt.Println("Data:")
		//fmt.Println(data)

		err = binary.Read(buffer, binary.BigEndian, bitacora) //* //& //NADA [la direcion de la otra var]][ir a buscarla]
		if err != nil {
			log.Fatal("binary.Read failed", err)
			file.Close()
			PrintError(ERROR, "Error al momento de decodificar la Bitacora")
			return false
		}

		//fmt.Println("Bitacora:")
		//fmt.Println(bitacora)
		return true

	}
	PrintError(comando, "Error al extraer la Bitacora del archivo")
	return false

}

//TOMAR EN CUENTA QUE ESTO PUEDE DEJAR MOVIDO EL PUNTERO DEL ARCHIVO

//ReadBytes retorna nil || []byte lee hasta n bytes CUIDADO con el puntero del archivo [seek(0,0)]
func ReadBytes(file *os.File, number int, lugarColocacion int, desde int, comando CONSTCOMANDO) []byte {
	// desde is the point of reference for offset
	// 0 = Beginning of file
	// 1 = Current position
	// 2 = End of file
	//--------------------
	// lugarColocacion is how many bytes to move
	// lugarColocacion can be positive or negative

	//*********OBTENER EL CURRENT POSITION DEL ARCHIVO*****************
	// Find the current position by getting the
	// return value from Seek after moving 0 bytes
	// currentPosition, err := file.Seek(0, 1)
	//	if err != nil {
	//    log.Fatal(err)
	//	}
	// fmt.Println("Current position:", currentPosition)
	//*****************************************************************

	newPosition, err := file.Seek(int64(lugarColocacion), desde)
	if err != nil {
		PrintError(ERROR, "Error al moverse a la posicion [Posicion:"+strconv.Itoa(lugarColocacion)+", Desde:"+strconv.Itoa(desde)+"] del disco para leer")
		log.Fatal(err)
		return nil
	}
	PrintAviso(comando, "Movidos a la posicion [Posicion:"+strconv.FormatInt(newPosition, 10)+", Desde:"+strconv.Itoa(desde)+"] para leer en el disco")

	bytesSlice := make([]byte, number)
	bytesRead, err := file.Read(bytesSlice) //LEER HASTA n BYTES DEL ARCHIVO
	if err != nil {
		PrintError(ERROR, "Error al leer en el archivo")
		log.Fatal(err)
		return nil
	}
	//fmt.Println("")
	log.Printf("Number of bytes read: %d\n", bytesRead)
	//fmt.Println("")
	//log.Printf("Data read: %s\n", bytesSlice)
	//fmt.Println("")
	return bytesSlice
}

//WriteBytes2 retorna true si logro escribir sin ningun problema o false si existio un problema
func WriteBytes2(file *os.File, bytes []byte, lugarColocacion int, desde int, comando CONSTCOMANDO) bool {

	nuevaPos, err := file.Seek(int64(lugarColocacion), desde)
	if err != nil {
		PrintError(ERROR, "Error al moverse a la posicion [Posicion:"+strconv.Itoa(lugarColocacion)+", Desde:"+strconv.Itoa(desde)+"] del disco para escribir")
		log.Fatal(err)
		return false
	}
	PrintAviso(comando, "Movidos a la posicion [Posicion:"+strconv.FormatInt(nuevaPos, 10)+", Desde:"+strconv.Itoa(desde)+"] para escribir en el disco")

	if _, err := file.Write(bytes); err != nil {
		PrintError(ERROR, "Error al escribir en el archivo")
		log.Fatal(err)
		return false
	}
	PrintAviso(comando, "Se escribio correctamente en el disco")
	return true
}

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

//ComandoExec ejecuta el comando Exec
func ComandoExec(arreglo []string) {
	for _, val := range arreglo {
		//fmt.Println(strconv.Itoa(i) + " = " + val)
		fmt.Println("======================================")
		fmt.Println("Please enter a command:")
		fmt.Println(val)
		fmt.Println("======================================")
		Separar(val)
	}
}

//ParametroPathExec evalua el parametro path para ejecutar el comando exec
func ParametroPathExec(comando CONSTCOMANDO, mapa map[string]string) (bool, []string) {

	if val, ok := mapa["PATH"]; ok {
		val = filepath.Join(val)

		if ExisteDirOrFile(val) {

			if filepath.Ext(val) == ".mia" {

				// Read file to byte slice
				data, err := ioutil.ReadFile(val)
				if err != nil {
					PrintError(comando, "Error, Al leer el archivo en la ruta ["+val+"]")
					log.Fatal(err)
					return false, nil
				}

				//log.Printf("Data read: %s\n", data)
				//strings.TrimRight(entrada, "\r\n\t")

				cadenaEntera := string(data)

				splSalto := strings.Split(cadenaEntera, "\n")

				comodin := ""
				sliceContendero := make([]string, 0)

				for _, val := range splSalto {

					if val != "" {
						if strings.Contains(val, "\\*") {

							comodin = comodin + strings.TrimSpace(val)
							continue

						} else {
							comodin = comodin + strings.TrimSpace(val)
							sliceContendero = append(sliceContendero, comodin)
							comodin = ""
						}
					}

				}

				//for i, val := range sliceContendero {
				//	fmt.Println(strconv.Itoa(i) + " = " + val)
				//}
				PrintAviso(comando, "Archivo leido, Ejecutando...")
				return true, sliceContendero

			}

			PrintError(comando, "El archivo en la ruta ["+val+"] no es de extension .mia no se puede leer")
			return false, nil

		}

		PrintError(comando, "El archivo en la ruta ["+val+"] no existe, no se puede leer")
		return false, nil

	}
	PrintError(comando, "El parametro obligatorio path no esta en la sentencia")
	return false, nil

}

//ComandoMkdisk ejecuta Mkdisk
func ComandoMkdisk(comando CONSTCOMANDO, mapa map[string]string, size int) {

	//FIXME: ATRIBUTOS DEL STRUCT MAYUSCUAL-MINUSCULA     https://stackoverflow.com/questions/34078427/how-to-read-packed-binary-data-in-go
	//FIXME: thing := binData{}
	//FIXME: binary.Read(file, binary.LittleEndian, &thing)   https://stackoverflow.com/questions/42462869/go-reading-binary-file-with-struct
	//FIXME: fmt.Println(thing)
	//FIXME:  defina una estructura para almacenar todos los atributos analizados: https://www.jonathan-petitcolas.com/2014/09/25/parsing-binary-files-in-go.html
	//FIXME: UserDataMaxSize uint32  (struct)-> el tipo del atributo debe de coincidir con los bytes que estoy leyendo
	//FIXME: devdungeon.com/content/working-files-go#write_buffered

	tamanio := TamanioTotal(comando, size, mapa["UNIT"])
	path := mapa["PATH"]
	nombre := mapa["NAME"]

	if tamanio != -1 {

		var mbrComodin MBRStruct
		if tamanio > binary.Size(mbrComodin) {
			pathCompleto := path + nombre
			if CrearArchivo(comando, pathCompleto, tamanio) {
				PrintAviso(comando, "Se creo el disco correctamente")
			} else {
				PrintError(ERROR, "No se pudo crear el disco")
				return
			}

		} else {
			PrintError(ERROR, "El tamanio del disco es menor al tamanio del MBR, por lo tanto no se puede crear el disco")
			return
		}

	} else {
		PrintError(comando, "Error al calcular el size en bytes")
		return
	}

}

//CrearArchivo plano con 0
func CrearArchivo(comando CONSTCOMANDO, path string, tamanio int) bool {

	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		PrintError(ERROR, "Error al crear el archivo")
		log.Fatal(err)
		return false
	}
	log.Println(file)

	//------------------------------------------------------------------------------------------binario

	relleno := make([]byte, tamanio)
	rellenoP := &relleno

	bufferBinario := new(bytes.Buffer)                                       //Un buffer de bytes nuevo
	if binary.Write(bufferBinario, binary.BigEndian, rellenoP); err != nil { //escritor binario //escribe la representacion binaria de MBR y lo almacena en el BUFFER
		PrintError(ERROR, "Error al hacer la conversion a binario")
		fmt.Println("binary.Write failed:", err)
		file.Close()
		return false
	}
	//fmt.Printf("% x", bufferBinario.Bytes()) //buffer.Bytes() //"abcde"  //[97 98 99 100 101]

	//------------------------------------------------------------------------------------------[]byte normal
	if !writeBytes(file, bufferBinario.Bytes()) { //[]byte //len() //arreglo de la representacion binario del parametro pasado
		return false
	}

	nuevaPos, err := file.Seek(0, 0)
	if err != nil {
		PrintError(ERROR, "Error al moverse al inicio del archivo para escribir el MBR")
		log.Fatal(err)
		return false
	}
	PrintAviso(comando, "Movidos a la posicion ["+strconv.FormatInt(nuevaPos, 10)+"] para escribir en el disco")

	//TODO: VERIFICAR QUE NO SEA NECESARIO INICIALIZAR LAS PARTICIONES

	//----------------------------------------------------------------MBR
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	mbr := &MBRStruct{ //TODO: OBSERVAR EL &
		MbrTamanio:       int64(tamanio),
		MbrDiskSignature: int64(r.Int()),
		//Partition:
	}
	copy(mbr.MbrFechaCreacion[:], time.Now().Format("01-02-2006 15:04:05"))

	//----------------------------------------------------------------PARTITION
	for i := 0; i < 4; i++ {
		mbr.Partition[i].PartStatus = '0'
		mbr.Partition[i].PartType = '0'
		mbr.Partition[i].PartFit = '0'
		mbr.Partition[i].PartStart = -1
		mbr.Partition[i].PartSize = 0
		for n := range mbr.Partition[i].PartName {
			mbr.Partition[i].PartName[n] = '0'
		}
	}
	//----------------------------------------------------------------

	bufferBinarioMBR := new(bytes.Buffer)                                  //Un buffer de bytes nuevo
	if binary.Write(bufferBinarioMBR, binary.BigEndian, mbr); err != nil { //escritor binario //escribe la representacion binaria de MBR y lo almacena en el BUFFER
		PrintError(ERROR, "Error al hacer la conversion a binario")
		fmt.Println("binary.Write failed:", err)
		file.Close()
		return false
	}
	//fmt.Printf("% x", bufferBinario.Bytes()) //buffer.Bytes() //"abcde"  //[97 98 99 100 101]

	//------------------------------------------------------------------------------------------[]byte normal
	if writeBytes(file, bufferBinarioMBR.Bytes()) { //[]byte //len() //arreglo de la representacion binario del parametro pasado
		return true
	}
	return false

}

func writeBytes(file *os.File, bytes []byte) bool { //escribir en el archivo el []byte completo
	if _, err := file.Write(bytes); err != nil {
		PrintError(ERROR, "Error al escribir en el archivo")
		log.Fatal(err)
		return false
	}
	return true
}

//Buffer buffer para crear el archivo
//func Buffer(size int) []byte {
//
//}

//TamanioTotal valor en bytes [size * unit] retorna -1 si hay error
func TamanioTotal(comando CONSTCOMANDO, size int, unit string) int {
	switch unit {
	case "B":
		return size
	case "K":
		return size * 1024
	case "M":
		return size * 1024 * 1024
	default:
		PrintError(comando, "Error al calcular el tamanio total")
		return -1
	}
}

//ComandoRmdisk ejecuta RMDISK
func ComandoRmdisk(comando CONSTCOMANDO, mapa map[string]string) {
	//Base extrae disco1.disk
	//Dir extrae la ruta de los directorios sin el nombre del archivo
	//Ext extrae la extension .disk
	//join para el caracter '/' y colocarlo en el caracter correcto para el SO
	//Split devuelve la ruta en un string y el archivo en otro string MAGNIFICO!
	//Walk camina sobre los directorios hasta llegar al archivo
	if val, ok := mapa["PATH"]; ok {
		val = filepath.Join(val)

		if ExisteDirOrFile(val) {
			PrintAviso(comando, "Se eliminara: ["+val+"] Desea proceder? [y/n]")
			confirmacion := CadenaConsola()
			if confirmacion == "y" {

				err := os.Remove(val)
				if err != nil {
					PrintError(comando, "Error al eliminar el disco ["+val+"] error:")
					fmt.Println(err)
					return
				}

				PrintAviso(comando, "Eliminacion exitosa de ["+val+"]")

			} else {
				PrintAviso(comando, "eliminacion abortada")
				return
			}

		} else {
			PrintError(comando, "El disco en la ruta ["+val+"] no existe, no se puede eliminar")
			return
		}

	} else {
		PrintError(comando, "El parametro obligatorio path no esta en la sentencia")
		return
	}
}

//ParametroUnit scanner para -unit
func ParametroUnit(comando CONSTCOMANDO, mapa map[string]string) int {
	if val, ok := mapa["UNIT"]; ok {
		val = strings.ToUpper(val)

		if val == "K" || val == "M" {
			if val == "K" {
				mapa["UNIT"] = "K"
				return 0
			} else if val == "M" {
				mapa["UNIT"] = "M"
				return 0
			} else {
				return -1
			}
		} else {
			PrintError(comando, "El valor del parametro unit no es correcto ["+val+"]")
			return -1
		}

	} else {
		PrintAviso(comando, "El parametro opcional unit no esta en la sentencia...")
		PrintAviso(comando, "Se le asignara el valor predefinido [M] Megabytes")
		mapa["UNIT"] = "M"
		return 0
	}
}

//ParametroUnitFdisk scanner para -unit
func ParametroUnitFdisk(comando CONSTCOMANDO, mapa map[string]string) int {
	if val, ok := mapa["UNIT"]; ok {
		val = strings.ToUpper(val)

		if val == "K" || val == "M" || val == "B" {
			if val == "K" {
				mapa["UNIT"] = "K"
				return 0
			} else if val == "M" {
				mapa["UNIT"] = "M"
				return 0
			} else if val == "B" {
				mapa["UNIT"] = "B"
				return 0
			} else {
				PrintError(comando, "El valor del parametro unit no es correcto ["+val+"]")
				return -1
			}
		} else {
			PrintError(comando, "El valor del parametro unit no es correcto ["+val+"]")
			return -1
		}

	} else {
		PrintAviso(comando, "El parametro opcional unit no esta en la sentencia...")
		PrintAviso(comando, "Se le asignara el valor predefinido [K] Kilobytes")
		mapa["UNIT"] = "K"
		return 0
	}
}

//ParametroTypeFdisk scanner para -type
func ParametroTypeFdisk(comando CONSTCOMANDO, mapa map[string]string) int {
	if val, ok := mapa["TYPE"]; ok {
		val = strings.ToUpper(val)

		if val == "P" || val == "E" || val == "L" {
			if val == "P" {
				mapa["TYPE"] = "P"
				return 0
			} else if val == "E" {
				mapa["TYPE"] = "E"
				return 0
			} else if val == "L" {
				mapa["TYPE"] = "L"
				return 0
			} else {
				PrintError(comando, "El valor del parametro type no es correcto ["+val+"]")
				return -1
			}
		} else {
			PrintError(comando, "El valor del parametro type no es correcto ["+val+"]")
			return -1
		}

	} else {
		PrintAviso(comando, "El parametro opcional type no esta en la sentencia...")
		PrintAviso(comando, "Se le asignara el valor predefinido [P] Primaria")
		mapa["TYPE"] = "P"
		return 0
	}
}

//ParametroFitFdisk scanner para -fit
func ParametroFitFdisk(comando CONSTCOMANDO, mapa map[string]string) int {
	if val, ok := mapa["FIT"]; ok {
		val = strings.ToUpper(val)

		if val == "BF" || val == "FF" || val == "WF" {
			if val == "BF" {
				mapa["FIT"] = "BF"
				return 0
			} else if val == "FF" {
				mapa["FIT"] = "FF"
				return 0
			} else if val == "WF" {
				mapa["FIT"] = "WF"
				return 0
			} else {
				PrintError(comando, "El valor del parametro fit no es correcto ["+val+"]")
				return -1
			}
		} else {
			PrintError(comando, "El valor del parametro fit no es correcto ["+val+"]")
			return -1
		}

	} else {
		PrintAviso(comando, "El parametro opcional fit no esta en la sentencia...")
		PrintAviso(comando, "Se le asignara el valor predefinido [WF] Peor Ajuste")
		mapa["FIT"] = "WF"
		return 0
	}
}

//ParametroDeleteFdisk scanner para -delete no viene, sin errores[false,0] viene, ok[true,0] viene, error[true,-1]
func ParametroDeleteFdisk(comando CONSTCOMANDO, mapa map[string]string) (bool, int) {
	if val, ok := mapa["DELETE"]; ok {
		val = strings.ToUpper(val)

		if val == "FAST" || val == "FULL" {
			if val == "FAST" {
				mapa["DELETE"] = "FAST"
				return true, 0
			} else if val == "FULL" {
				mapa["DELETE"] = "FULL"
				return true, 0
			} else {
				PrintError(comando, "El valor del parametro delete no es correcto ["+val+"]")
				return true, -1
			}
		} else {
			PrintError(comando, "El valor del parametro delete no es correcto ["+val+"]")
			return true, -1
		}

	} else {
		return false, 0
	}
}

//ParametroNameFdisk scanner para -name
func ParametroNameFdisk(comando CONSTCOMANDO, mapa map[string]string) int {
	if val, ok := mapa["NAME"]; ok {

		if val != "" {
			return 0
		}

		PrintError(comando, "El valor del parametro name es vacio ["+val+"]")
		return -1

	}
	PrintAviso(comando, "El parametro obligatorio name no esta en la sentencia...")
	return -1

}

//ParametroSize scanner para -size [-1 si hay problema]
func ParametroSize(comando CONSTCOMANDO, mapa map[string]string) int {
	if valMapa, ok := mapa["SIZE"]; ok {
		if val, bol := EsNumero(valMapa); (bol) && (val > 0) {
			return val
		}
		PrintError(comando, "El parametro obligatorio size no es de tipo numerico o no es > 0.")
		return -1

	}
	PrintError(comando, "El parametro obligatorio size no esta en la sentencia")
	return -1
}

//ParametroSizeMkfile scanner para -size [-1 si hay problema]
func ParametroSizeMkfile(comando CONSTCOMANDO, mapa map[string]string) (bool, int) {
	if valMapa, ok := mapa["SIZE"]; ok {
		if val, bol := EsNumero(valMapa); (bol) && (val >= 0) {
			return true, val
		}
		PrintError(comando, "El parametro opcional size no es de tipo numerico o no es >= 0. Posiblemente sea un numero negativo")
		return false, -1
	}
	PrintError(comando, "El parametro opcional size no esta en la sentencia")
	return false, -1
}

//ParametroAddFdisk verifica que venga y que sea numero retorna el valor [int]
func ParametroAddFdisk(comando CONSTCOMANDO, mapa map[string]string) (bool, int) {
	if valMapa, ok := mapa["ADD"]; ok {
		if val, bol := EsNumero(valMapa); bol {
			return true, val
		}
		PrintError(comando, "El parametro opcional add no es de tipo numerico.")
		return false, -1

	}
	return false, 0
}

//ParametroPathMkdisk scanner -path [MKDISK]
func ParametroPathMkdisk(comando CONSTCOMANDO, mapa map[string]string) int {

	//Base extrae disco1.disk
	//Dir extrae la ruta de los directorios sin el nombre del archivo
	//Ext extrae la extension .disk
	//join para el caracter '/' y colocarlo en el caracter correcto para el SO
	//Split devuelve la ruta en un string y el archivo en otro string MAGNIFICO!
	//Walk camina sobre los directorios hasta llegar al archivo
	if val, ok := mapa["PATH"]; ok {
		val = filepath.Join(val)

		if !ExisteDirOrFile(val) {
			PrintAviso(comando, "Se creara la ruta :"+val)
			err := os.MkdirAll(val, os.ModePerm)
			if err != nil {
				PrintError(comando, "Problemas al crear la ruta "+val)
				fmt.Println(err)
				return -1
			}
			PrintAviso(comando, "Ruta creada: "+val)
			return 0
		}
		PrintAviso(comando, "La ruta ["+val+"] ya existe, se procede")
		return 0

	}
	PrintError(comando, "El parametro obligatorio path no esta en la sentencia")
	return -1

}

//ParametroPathFdisk scanner -path [FDISK]
func ParametroPathFdisk(comando CONSTCOMANDO, mapa map[string]string) int {

	if val, ok := mapa["PATH"]; ok {
		val = filepath.Join(val)

		if ExisteDirOrFile(val) {
			PrintAviso(comando, "Si existe el disco, se procede")
			return 0
		}

		PrintError(comando, "El disco en la ruta ["+val+"] no existe, no se puede crear una particion")
		return -1

	}
	PrintError(comando, "El parametro obligatorio path no esta en la sentencia")
	return -1

}

//ParametroPathMount scanner -path [MOUNT]
func ParametroPathMount(comando CONSTCOMANDO, mapa map[string]string) int {

	if val, ok := mapa["PATH"]; ok {
		val = filepath.Join(val)

		if ExisteDirOrFile(val) {
			PrintAviso(comando, "Si existe el disco, se procede")
			return 0
		}

		PrintError(comando, "El disco en la ruta ["+val+"] no existe, no se puede montar una particion")
		return -1

	}
	PrintError(comando, "El parametro obligatorio path no esta en la sentencia")
	return -1

}

//ParametroPathRep crea las carpetas a partir del path cortado hasta el nombre
func ParametroPathRep(comando CONSTCOMANDO, mapa map[string]string) (bool, string) { //path del reporte

	//"/home/user/reports/reporte 2.pdf"

	//Base extrae disco1.disk
	//Dir extrae la ruta de los directorios sin el nombre del archivo
	//Ext extrae la extension .disk
	//join para el caracter '/' y colocarlo en el caracter correcto para el SO
	//Split devuelve la ruta en un string y el archivo en otro string MAGNIFICO!
	//Walk camina sobre los directorios hasta llegar al archivo
	if val, ok := mapa["PATH"]; ok {

		dir, _ := filepath.Split(val)

		dir = filepath.Join(dir)

		if !ExisteDirOrFile(dir) {
			PrintAviso(comando, "Se creara la ruta :"+dir)
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				PrintError(comando, "Problemas al crear la ruta "+dir)
				fmt.Println(err)
				return false, ""
			}
			PrintAviso(comando, "Ruta creada: "+dir)
			return true, val
		}
		PrintAviso(comando, "La ruta ["+dir+"] ya existe, se procede")
		return true, val

	}
	PrintError(comando, "El parametro obligatorio path no esta en la sentencia")
	return false, ""

}

//ParametroNameMkdisk scanner -name para [Mkdisk]
func ParametroNameMkdisk(comando CONSTCOMANDO, mapa map[string]string) int {
	if val, ok := mapa["NAME"]; ok { //QUE TENGA LA EXTENSION

		if strings.Contains(val, ".") {
			extension := strings.Split(val, ".")
			if len(extension) == 2 {
				if extension[1] != "" {
					extension[1] = strings.ToUpper(extension[1])
					if !strings.HasSuffix(extension[1], "DSK") {
						PrintError(comando, "EL nombre no tiene una extension correcta, favor verificar")
						return -1
					}
				} else {
					PrintError(comando, "EL nombre no tiene una extension correcta, favor verificar")
					return -1
				}
			} else {
				PrintError(comando, "EL nombre no tiene una extension correcta, favor verificar")
				return -1
			}
		} else {
			PrintError(comando, "EL nombre no tiene una extension correcta, favor verificar")
			return -1
		}

		if pat, ok := mapa["PATH"]; ok { //TRAER LA RUTA
			pat = filepath.Join(pat)
			pathCompleto := pat + "/" + val
			if !ExisteDirOrFile(pathCompleto) { //QUE NO EXISTA LA RUTA COMPLETA CON EL NOMBRE DEL DISCO
				//TODO: creo el archivo
				//fmt.Println("lista para crear el archivo " + pathCompleto)
				return 0
			}
			PrintError(comando, "El disco ["+pathCompleto+"] ya existe en la direccion")
			return -1

		}
		PrintError(comando, "No se puede crear el disco debido a problemas con el path")
		return -1

	}
	PrintError(comando, "El parametro obligatorio name no esta en la sentencia")
	return -1

}

//EsPar true si es impar el numero
func EsPar(i int) bool {
	i = i + 1
	if (i % 2) == 0 {
		return true
	}
	return false
}

//EsNumero retorna un int, bool dice si es numero la cadena y te devuelve el valor int [-1, falso si hay error]
func EsNumero(cadena string) (int, bool) {
	cadena = strings.TrimSpace(cadena)
	if val, err := strconv.Atoi(cadena); err == nil {
		return val, true
	}
	return -1, false
}

//ExisteDirOrFile true si existe el directorio o archivo
func ExisteDirOrFile(pathFil string) bool {
	if _, err := os.Stat(pathFil); !os.IsNotExist(err) {
		// path/to/whatever exists
		return true
	}
	return false
}

////PrintSlice imprime un slice
//func PrintSlice(name string, x []Comando) {
//	fmt.Printf("%s len=%d cap=%d %v\n",
//		name, len(x), cap(x), x)
//}

//PrintAviso imprime un aviso
func PrintAviso(comando CONSTCOMANDO, mensaje string) {
	fmt.Println("AVISO! [" + comando.String() + "]->  " + mensaje)
}

//PrintAviso2 imprime un aviso
func PrintAviso2(comando string, mensaje string) {
	fmt.Println("AVISO! [" + comando + "]->  " + mensaje)
}

//PrintError imprime un error
func PrintError(comando CONSTCOMANDO, mensaje string) {
	fmt.Println("ERROR! [" + comando.String() + "]->  " + mensaje)
}

//String para obtener el string del cons CONSTCOMANDO
func (d CONSTCOMANDO) String() string {
	return [...]string{"MKDISK", "RMDISK", "FDISK", "MOUNT", "UNMOUNT", "REP", "EXEC",
		"MKFS",
		"LOGIN",
		"LOGOUT",
		"MKGRP",
		"RMGRP",
		"MKUSR",
		"RMUSR",
		"CHMOD",
		"MKFILE",
		"CAT",
		"RM",
		"EDIT",
		"REN",
		"MKDIR",
		"CP",
		"MV",
		"FIND",
		"CHOWN",
		"CHGRP",
		"ERROR"}[d]
}

//Cons retorna un comando apartir de su representacion en string [-1 si hay error]
func Cons(comando string) CONSTCOMANDO {
	switch comando {
	case "MKDISK":
		return MKDISK
	case "RMDISK":
		return RMDISK
	case "FDISK":
		return FDISK
	case "MOUNT":
		return MOUNT
	case "UNMOUNT":
		return UNMOUNT
	case "REP":
		return REP
	case "EXEC":
		return EXEC
	case "MKFS":
		return MKFS
	case "LOGIN":
		return LOGIN
	case "LOGOUT":
		return LOGOUT
	case "MKGRP":
		return MKGRP
	case "RMGRP":
		return RMGRP
	case "MKUSR":
		return MKUSR
	case "RMUSR":
		return RMUSR
	case "CHMOD":
		return CHMOD
	case "MKFILE":
		return MKFILE
	case "CAT":
		return CAT
	case "RM":
		return RM
	case "EDIT":
		return EDIT
	case "REN":
		return REN
	case "MKDIR":
		return MKDIR
	case "CP":
		return CP
	case "MV":
		return MV
	case "FIND":
		return FIND
	case "CHOWN":
		return CHOWN
	case "CHGRP":
		return CHGRP
	default:
		return ERROR
	}
}

/*
go build ...nombredelalibreria (ubicado en la carpeta /bin) (cd lib)
go install (ubicado en la carpeta /bin)
go build  (ubicado en la carpeta del proyecto SOBRE EL EJECUTABLE) (cd ..)
ls
ls -la
./Proyecto1
go run main.go
*/

/*
	s := []int{2, 3, 5}
	printSlice(s)
	s = append(s, 6)
	printSlice(s)

	a := make([]int, 3)
	printSlice(a)
	for i, val := range a{
		if val == 0{
			a[i] = 6
		}
	}

	a = append(a, 7)
	printSlice(a)
	=============================
	len=3 cap=3 [2 3 5]
	len=4 cap=6 [2 3 5 6]
	len=3 cap=3 [0 0 0]
	len=4 cap=6 [6 6 6 7]


[_a-zA-Z0-9\/.]+ 		   (cualquier cadena, path sin "", id)
^"[_a-zA-Z0-9\/. ]+["$]    (path con "*******")

*/
