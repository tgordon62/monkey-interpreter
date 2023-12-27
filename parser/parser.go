// Pratt parser for the Monkey programming language.
package parser

import (
	"fmt"
	"monkey-interpreter/ast"
	"monkey-interpreter/lexer"
	"monkey-interpreter/token"
	"strconv"
)

// Structure of a parser instance.
type Parser struct {
	lex       *lexer.Lexer
	errors    []string
	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// Function types which correspond to token positions.
type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

// Create a new parser instance and reutrn it. Accepts a Lexer instance
// which will be parsed.
func New(lex *lexer.Lexer) *Parser {
	par := &Parser{
		lex:    lex,
		errors: []string{},
	}

	// Add parsing functions to maps
	par.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	par.registerPrefix(token.IDENT, par.parseIdentifier)
	par.registerPrefix(token.INT, par.parseIntegerLiteral)

	// Read two tokens, so curToken and peekToken are both set
	par.nextToken()
	par.nextToken()

	return par
}

// Parse the tokens generated by the Lexer.
func (par *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !par.curTokenIs(token.EOF) {
		stmt := par.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		par.nextToken()
	}

	return program
}

// Parse an entire statement beginning with either a 'let' or 'return' keyword.
func (par *Parser) parseStatement() ast.Statement {
	switch par.curToken.Type {
	case token.LET:
		return par.parseLetStatement()
	case token.RETURN:
		return par.parseReturnStatement()
	default:
		return par.parseExpressionStatement()
	}
}

// Parse a statement beginning with a 'let' keyword.
func (par *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: par.curToken}

	if !par.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: par.curToken, Value: par.curToken.Literal}

	if !par.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO: We're skipping the expressions until we
	// encounter a semicolon
	for !par.curTokenIs(token.SEMICOLON) {
		par.nextToken()
	}

	return stmt
}

// Parse a statement beginning with a 'return' keyword.
func (par *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: par.curToken}

	par.nextToken()

	// TODO: We're skipping the expressions until we
	// encounter a semicolon
	for !par.curTokenIs(token.SEMICOLON) {
		par.nextToken()
	}

	return stmt
}

// Parse an expression statement.
func (par *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: par.curToken}
	stmt.Expression = par.parseExpression(LOWEST)

	if par.peekTokenIs(token.SEMICOLON) {
		par.nextToken()
	}

	return stmt
}

// Parse an expression.
func (par *Parser) parseExpression(precedence int) ast.Expression {
	prefix := par.prefixParseFns[par.curToken.Type]
	if prefix == nil {
		return nil
	}
	leftExp := prefix()

	return leftExp
}

// Parse an identifier.
func (par *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: par.curToken, Value: par.curToken.Literal}
}

// Parse an integer literal.
func (par *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: par.curToken}

	value, err := strconv.ParseInt(par.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", par.curToken.Literal)
		par.errors = append(par.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

// Update the current token of the parser instace to the peek token, and then
// update the peek token to the next token from the Lexer.
func (par *Parser) nextToken() {
	par.curToken = par.peekToken
	par.peekToken = par.lex.NextToken()
}

// Confirm that the Parser's current token is the same type as the parameter tok.
func (par *Parser) curTokenIs(tok token.TokenType) bool {
	return par.curToken.Type == tok
}

// Confirm that the Parser's peek token is the same type as the parameter tok.
func (par *Parser) peekTokenIs(tok token.TokenType) bool {
	return par.peekToken.Type == tok
}

// Confirm that the Parser's peek token is the same type as the parameter tok.
// If it is, then the parser will proceed to the next token. If it is not, the
// function will log an error.
func (par *Parser) expectPeek(tok token.TokenType) bool {
	if par.peekTokenIs(tok) {
		par.nextToken()
		return true
	} else {
		par.peekError(tok)
		return false
	}
}

// Add a function for parsing prefix operators to the prefixParseFns map.
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// Add a function for parsing infix operators to the infixParseFns map.
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// Return the log of errors generated while parsing.
func (par *Parser) Errors() []string {
	return par.errors
}

// Log an error signifyiing that the expected next token was not found.
func (par *Parser) peekError(tok token.TokenType) {
	msg := fmt.Sprintf("Expected next token to be %s, got %s instead",
		tok, par.peekToken.Type)
	par.errors = append(par.errors, msg)
}
