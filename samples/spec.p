# example

VAR msg

msg = "hello world"

CLEAR
LOCATE 15, 30
COLOR "green"
PRINT msg
PIXEL 30, 20
LINE 30, 20, 30, 40
CIRCLE 20, 4, 3

## language

VAR <var>
IF <exp> {

}
ELSE {

}

var = val

LOOP {
  BREAK
  NEXT
}

WHILE <expr> {
	BREAK
	NEXT
}

IMPORT <mod>
IMPORT <mod> WITHPREFIX x.
UNIMPORT <mod>
UNDEFINE <var>, <var>
EXPORT <var>, <var>

FUNC name(<var>, <var>) {

  RETURN <expr>
}

PROC name var, var {
  DONE
}

CALL func()

strings
integers
bools
AND
OR
parenthesis
==
!=
<
<=
>
>=

var[index]

# standard module

CLEAR
LOCATE x, y
COLOR color
PRINT msg
PIXEL x, y
LINE x1, y1, x2, y2
CIRCLE x, y, r

KEYBOARD function
KEYBOARDOFF
MOUSE function, function
MOUSEOFF
WAIT
INPUT()

# console module

CLEAR
LOCATE x, y
COLOR color
PRINT msg
INPUT()

# datastructure module

LIST()
MAP()
APPEND list, val
POP list
LEN(list)
DELETE map, index
